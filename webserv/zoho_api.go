package webserv

import (
	"bytes"
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv/gen/webserv_data"
	"io"
	"net/http"
	"time"
)

type ZohoStruct struct {
	config     ZohoConfigBuilder
	modified   bool
	bodyMap    JSMap
	bodyBytes  []byte
	queryParam []string
	fatalError error
}

type Zoho = *ZohoStruct

func PrepareZoho(config ZohoConfig) {
	CheckState(sharedZoho == nil)
	sharedZoho = &ZohoStruct{}
	sharedZoho.initConfig(config)
}

func SharedZoho() Zoho {
	CheckState(sharedZoho != nil)
	return sharedZoho
}

func (z Zoho) initConfig(config ZohoConfig) {
	if config == nil {
		config = DefaultZohoConfig
	}
	f := z.cacheFile()
	z.modified = true
	if f.Exists() {
		config = config.Parse(JSMapFromFileM(f)).(ZohoConfig)
		z.modified = false
	}
	z.config = config.ToBuilder()
}

func (z Zoho) cacheFile() Path {
	return NewPathM(".zoho_config.json")
}

func (z Zoho) RefreshToken() string {
	tokn := z.config.RefreshToken()
	if tokn == "" {
		// [] Open a browser, and go to:  https://api-console.zoho.com/
		// [] Select (or create, if necessary) the 'Self Client' application
		// [] For Scope, paste this (no spaces):
		//    ZohoMail.messages.ALL,ZohoMail.attachments.ALL,ZohoMail.folders.ALL,ZohoMail.accounts.ALL
		// [] Choose duration 10 minutes
		// [] Enter 'all' in the scope description
		// [] Press 'Create'
		// [] Paste the result into the following curl command, where it says 'code':
		//
		// **NOTE** The client_id and client_secret are to be copied from the 'Self Client' Client Secret tab,
		//           NOT the 'Animal' Client Secret tab!
		//
		// curl https://accounts.zoho.com/oauth/v2/token \
		// -X POST \
		// -d "client_id=........" \
		// -d "client_secret=............"\
		// -d "code=...paste the code here..."\
		// -d "grant_type=authorization_code"

		// It returns something like this:
		//
		// {
		//  "access_token":"....",
		//  "refresh_token":"....",
		//   "api_domain":"https://www.zohoapis.com",
		//   "token_type":"Bearer",
		//   "expires_in":3600
		// }
		BadState("there is no refresh token")
	}
	return tokn
}

func (z Zoho) AccessToken() string {
	c := z.config
	if c.AccessToken() == "" || c.AccessTokenExpiryMs() < time.Now().UnixMilli() {
		pr := PrIf("zoho_api;AccessToken", false)
		pr("using refresh token to get fresh access token")
		// Request a new access token using the refresh token
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodPost, "https://accounts.zoho.com/oauth/v2/token", nil)
		CheckOk(err)

		p := req.URL.Query()

		p.Add("refresh_token", z.RefreshToken())
		p.Add("grant_type", "refresh_token")
		p.Add("client_id", c.ClientId())
		p.Add("client_secret", c.ClientSecret())
		p.Add("redirect_uri", "https://pawsforaid.org/zoho")

		req.URL.RawQuery = p.Encode()
		pr("Encoded URL:", req)

		req.Header.Add("Accept", "application/json")
		resp, err := client.Do(req)
		CheckOk(err)

		defer resp.Body.Close()
		responseBody, err := io.ReadAll(resp.Body)
		CheckOk(err)

		pr("resp.Status:", resp.Status)
		pr("resp.Body:", INDENT, string(responseBody))

		mp, err := JSMapFromString(string(responseBody))
		if err != nil {
			BadState("failed to parse JSMap from response body:", INDENT, string(responseBody))
		}
		z.editConfig().SetAccessToken(mp.GetString("access_token")).
			SetAccessTokenExpiryMs(time.Now().UnixMilli() + (mp.GetInt64("expires_in")-120)*1000)
		z.flushConfig()
	}
	return c.AccessToken()
}

func (z Zoho) AccountId() string {
	pr := PrIf("AccountId", false)
	if z.config.AccountId() == "" || z.config.FromAddress() == "" {
		accId := "?"
		fromAddr := "?"

		mp, err := z.makeAPICallJson()
		if !z.setFatalErrorIf(err) {
			data := mp.GetList("data").Get(0).AsJSMap()
			accId = data.GetString("accountId")
			fromAddr = data.GetString("incomingUserName")
		}
		z.editConfig().SetAccountId(accId).SetFromAddress(fromAddr)
		z.flushConfig()
		pr("account id:", accId)
	}
	return z.config.AccountId()
}

func (z Zoho) FromAddress() string {
	z.AccountId()
	return z.config.FromAddress()
}

func (z Zoho) editConfig() ZohoConfigBuilder {
	z.modified = true
	return z.config
}

func (z Zoho) flushConfig() {
	if z.modified {
		pr := PrIf("flushConfig", true)
		z.modified = false
		f := z.cacheFile()
		f.WriteStringM(z.config.String())
		pr("flushed:", INDENT, z.config)
	}
}

var sharedZoho Zoho

func (z Zoho) setFatalErrorIf(err error) bool {
	if err != nil && z.fatalError == nil {
		z.fatalError = err
		return true
	}
	return false
}
func (z Zoho) Folders() (map[string]string, error) {
	if z.fatalError == nil && len(z.config.FolderMap()) == 0 {
		pr := PrIf("Folders", true)
		accountId := z.AccountId()
		mp, err := z.makeAPICallJson(accountId, "folders")
		if err != nil {
			return nil, err
		}
		pr("results:", INDENT, mp)

		jl := mp.GetList("data")
		var x = make(map[string]string)
		for _, m := range jl.AsMaps() {
			pr("map:", INDENT, m)
			// The folderId are appearing as *strings* in the jsmap from zoho, but elsewhere in their
			// API they want them sent as integers.
			x[m.GetString("folderName")] = m.GetString("folderId")
			//x[m.GetString("folderName")] = int64(ParseIntM(m.GetString("folderId")))
		}
		z.editConfig().SetFolderMap(x)
		z.flushConfig()
		pr("parsed:", INDENT, z.config.FolderMap())
	}
	return z.config.FolderMap(), nil
}

func (z Zoho) makeAPICallJson(args ...any) (JSMap, error) {
	bytes, err := z.makeAPICall(args...)
	if err != nil {
		return nil, err
	}
	mp, err := JSMapFromString(string(bytes))
	err = z.verifyOkCode(err, mp)
	if err != nil {
		mp = nil
	}
	return mp, err
}

func (z Zoho) verifyOkCode(err error, mp JSMap) error {
	if err != nil {
		return err
	}
	m2 := mp.OptMap("status")
	if !(m2 != nil && m2.OptInt("code", -1) == 200) {
		return EmailErrorAPIError
	}
	return nil
}

func (z Zoho) makeAPICall(args ...any) ([]byte, error) {
	pr := PrIf("makeAPICall", false)
	if z.fatalError != nil {
		return nil, z.fatalError
	}

	// copy some fields to locals and clear them immediately, in case there is some error later

	bodyBytes := z.bodyBytes
	z.bodyBytes = nil
	bodyMap := z.bodyMap
	z.bodyMap = nil
	queryParam := z.queryParam
	z.queryParam = nil

	// Body, if one is included, can be either bytes or JSMap, but not both
	CheckState(bodyBytes == nil || bodyMap == nil)

	method := http.MethodGet
	pr("args:", args)
	pr("body:", bodyMap)
	pr("queryParam:", queryParam)

	url := "https://mail.zoho.com/api/accounts"
	for _, x := range args {
		str := ToString(x)
		url = url + "/" + str
	}

	var body io.Reader

	if bodyMap != nil {
		bodyBytes = []byte(bodyMap.CompactString())
	}
	if bodyBytes != nil {
		method = http.MethodPost
		body = bytes.NewReader(bodyBytes)
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Zoho-oauthtoken "+z.AccessToken())

	if len(queryParam) != 0 {
		q := req.URL.Query()
		for i := 0; i < len(queryParam); i += 2 {
			q.Add(queryParam[i], queryParam[i+1])
		}
		// It seems I have to do this myself
		req.URL.RawQuery = q.Encode()
	}

	resp := CheckOkWith(client.Do(req))
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (z Zoho) body() JSMap {
	if z.bodyMap == nil {
		z.bodyMap = NewJSMap()
	}
	return z.bodyMap
}

func (z Zoho) addParam(key string, value string) {
	z.queryParam = append(z.queryParam, []string{key, value}...)
}

func (z Zoho) ReadInbox() ([]Email, error) {
	pr := PrIf("ReadInbox", false)
	f, err := z.Folders()
	if err != nil {
		return nil, err
	}

	id := f["Inbox"]
	// The parameters end up being strings anyways, so confusion about string vs int doesn't matter
	z.addParam("folderId", id)
	mp, err := z.makeAPICallJson(z.AccountId(), "messages", "view")
	if err != nil {
		return nil, err
	}

	data := mp.GetList("data")
	var results []Email

	for _, x := range data.AsMaps() {
		pr("Zoho:", INDENT, x)
		msgId := x.GetString("messageId")
		em := NewEmail()

		em.SetMessageId(msgId)
		em.SetFromAddress(x.GetString("sender"))
		em.SetToAddress(x.GetString("toAddress"))
		em.SetSubject(x.GetString("subject"))
		em.SetBody(x.GetString("summary"))
		em.SetReceivedTimeMs(int64(ParseInt64M(x.GetString("receivedTime"))))

		var atts []Attachment
		var totalAttSize int

		hasAttachment := x.GetString("hasAttachment") != "0"
		if hasAttachment {
			mp2, err := z.makeAPICallJson(z.AccountId(), "folders", id, "messages", msgId, "attachmentinfo")
			if err != nil {
				return nil, err
			}
			pr("attachment info:", mp2)
			alist := mp2.GetMap("data").GetList("attachments")
			for _, y := range alist.AsMaps() {
				att := NewAttachment()
				att.SetAttachmentId(y.GetString("attachmentId"))
				att.SetName(y.GetString("attachmentName"))
				size := y.GetInt("attachmentSize")
				totalAttSize += size
				if size > z.config.MaxAttachmentSize() || totalAttSize > z.config.MaxAttachmentTotalSize() {
					return nil, EmailErrorAttachmentSizeLimitExceeded
				}
				atts = append(atts, att.Build())
			}
		}

		//https://mail.zoho.com/api/accounts/<accountId>/folders/<folderId>/messages/<messageId>/attachments/<attachId>
		// Get the attachment data

		Todo("!Put limit on attachment size (and total size)", totalAttSize)
		for i, ati := range atts {
			ab := ati.ToBuilder()
			bytes, err := z.makeAPICall(z.AccountId(), "folders", id, "messages", msgId, "attachments", ati.AttachmentId())
			if err != nil {
				return nil, err
			}
			ab.SetData(bytes)
			atts[i] = ab.Build()
		}
		em.SetAttachments(atts)
		results = append(results, em.Build())
	}
	pr("results:", INDENT, results)
	return results, nil
}

func trim(s string) string {
	if len(s) > 40 {
		return s[0:40] + "..."
	}
	return s
}

func EmailSummary(e Email) JSMap {
	m := NewJSMap()
	m.Put("to", e.ToAddress())
	m.Put("subject", e.Subject())
	m.Put("body", trim(e.Body()))
	m.Put("# att", len(e.Attachments()))
	return m
}

var EmailErrorAttachmentSizeLimitExceeded = Error("attachment(s) size limit exceeded")
var EmailErrorAPIError = Error("Zoho API returned unexpected results")

func (z Zoho) SendEmail(email Email) error {

	// Uploading attachments:  https://www.zoho.com/mail/help/api/post-upload-attachments.html
	// Sending email: https://www.zoho.com/mail/help/api/post-send-an-email.html

	pr := PrIf("SendEmail", false)
	pr("email:", INDENT, EmailSummary(email))

	CheckArg(email.ToAddress() != "")
	CheckArg(email.Subject() != "")
	CheckArg(email.Body() != "")
	fromAddr := email.FromAddress()
	if fromAddr == "" {
		fromAddr = z.FromAddress()
	}
	CheckArg(fromAddr == z.FromAddress())

	total_size := 0
	max_size := 0
	for _, x := range email.Attachments() {
		s := len(x.Data())
		max_size = MaxInt(max_size, s)
		total_size += s
	}
	if max_size > z.config.MaxAttachmentSize() || total_size > z.config.MaxAttachmentTotalSize() {
		return EmailErrorAttachmentSizeLimitExceeded
	}

	// If there are attachment(s) to send, call the attachments api
	//https://mail.zoho.com/api/accounts/<accountId>/messages/attachments
	attachmentsList := NewJSList()
	if len(email.Attachments()) != 0 {
		for _, x := range email.Attachments() {
			pr("uploading attachment:", x.Name(), "length:", len(x.Data()))
			z.bodyBytes = x.Data()
			z.addParam("fileName", x.Name())
			z.addParam("isInline", "false")
			result, err := z.makeAPICallJson(z.AccountId(), "messages", "attachments")
			if err != nil {
				return err
			}
			attachmentsList.Add(result.GetMap("data"))
		}
	}
	pr("attachments list:", INDENT, attachmentsList)

	m := z.body()
	m.Put("fromAddress", fromAddr)
	m.Put("toAddress", email.ToAddress())
	m.Put("subject", email.Subject())
	m.Put("content", email.Body())
	m.Put("mailFormat", "plaintext")
	if attachmentsList.Length() != 0 {
		m.Put("attachments", attachmentsList)
	}
	result, err := z.makeAPICall(z.AccountId(), "messages")
	pr("result length:", len(result))
	return err
}
