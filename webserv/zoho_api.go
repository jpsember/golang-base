package webserv

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webserv/gen/webserv_data"
	"io"
	"net/http"
	"strings"
	"time"
)

type ZohoStruct struct {
	config     ZohoConfigBuilder
	modified   bool
	bodyMap    JSMap
	queryParam []string
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
		mp := z.makeAPICallJson()
		CheckState(mp.GetList("data").Length() == 1)
		data := mp.GetList("data").Get(0).AsJSMap()
		accountId := data.GetString("accountId")
		z.editConfig().SetAccountId(accountId).SetFromAddress(data.GetString("incomingUserName"))
		z.flushConfig()
		pr("account id:", accountId)
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

func (z Zoho) Folders() map[string]string {
	if len(z.config.FolderMap()) == 0 {
		pr := PrIf("Folders", true)
		mp := z.makeAPICallJson(z.AccountId(), "folders")
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
	return z.config.FolderMap()
}

func (z Zoho) makeAPICallJson(args ...any) JSMap {
	bytes := z.makeAPICall(args...)
	mp := JSMapFromStringM(string(bytes))
	if mp.GetMap("status").GetInt("code") != 200 {
		BadState("returned unexpected status:", INDENT, mp)
	}
	return mp
}

func (z Zoho) makeAPICall(args ...any) []byte {
	pr := PrIf("makeAPICall", false)

	// copy some fields to locals and clear them immediately, in case there is some error later
	b := z.bodyMap
	z.bodyMap = nil
	p := z.queryParam
	z.queryParam = nil

	method := http.MethodGet

	pr("args:", args)
	pr("body:", z.bodyMap)

	url := "https://mail.zoho.com/api/accounts"
	for _, x := range args {
		str := ToString(x)
		url = url + "/" + str
	}
	pr("url:", url)

	var body io.Reader

	if b != nil {
		Alert("Setting method=POST makes zoho complain")
		method = http.MethodPost
		pr("setting method=POST, body:", INDENT, b)
		body = io.NopCloser(strings.NewReader(b.CompactString()))
	}

	client := &http.Client{}
	req := CheckOkWith(http.NewRequest(method, url, body))
	req.Header.Set("Authorization", "Zoho-oauthtoken "+z.AccessToken())

	if len(p) != 0 {
		q := req.URL.Query()
		for i := 0; i < len(p); i += 2 {
			q.Add(p[i], p[i+1])
		}
	}

	resp := CheckOkWith(client.Do(req))
	defer resp.Body.Close()
	return CheckOkWith(io.ReadAll(resp.Body))
}

func (z Zoho) body() JSMap {
	if z.bodyMap == nil {
		z.bodyMap = NewJSMap()
		Pr("init body map to empty map", Callers(0, 5))
	}
	return z.bodyMap
}

func (z Zoho) addParam(key string, value string) {
	z.queryParam = append(z.queryParam, []string{key, value}...)
}

func (z Zoho) ReadInbox() []Email {
	pr := PrIf("ReadInbox", false)
	id := z.Folders()["Inbox"]
	// The parameters end up being strings anyways, so confusion about string vs int doesn't matter
	z.addParam("folderId", id)
	mp := z.makeAPICallJson(z.AccountId(), "messages", "view")

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
			mp2 := z.makeAPICallJson(z.AccountId(), "folders", id, "messages", msgId, "attachmentinfo")
			pr("attachment info:", mp2)
			alist := mp2.GetMap("data").GetList("attachments")
			for _, y := range alist.AsMaps() {
				att := NewAttachment()
				att.SetAttachmentId(y.GetString("attachmentId"))
				att.SetName(y.GetString("attachmentName"))
				att.SetSize(y.GetInt("attachmentSize"))
				totalAttSize += att.Size()
				atts = append(atts, att.Build())
			}
		}

		//https://mail.zoho.com/api/accounts/<accountId>/folders/<folderId>/messages/<messageId>/attachments/<attachId>
		// Get the attachment data
		Todo("!Put limit on attachment size (and total size)", totalAttSize)
		for i, ati := range atts {
			ab := ati.ToBuilder()
			bytes := z.makeAPICall(z.AccountId(), "folders", id, "messages", msgId, "attachments", ati.AttachmentId())
			CheckState(len(bytes) == ati.Size(), "size mismatch; expected", ati.Size(), "but got", len(bytes))
			ab.SetData(bytes)
			atts[i] = ab.Build()
		}
		em.SetAttachments(atts)
		results = append(results, em.Build())
	}
	pr("results:", INDENT, results)
	return results
}

func (z Zoho) SendEmail(email Email) {
	pr := PrIf("SendEmail", true)
	pr("email:", INDENT, email)

	CheckArg(email.ToAddress() != "")
	CheckArg(email.Subject() != "")
	CheckArg(email.Body() != "")
	fromAddr := email.FromAddress()
	if fromAddr == "" {
		fromAddr = z.FromAddress()
	}
	CheckArg(fromAddr == z.FromAddress())

	m := z.body()
	m.Put("fromAddress", fromAddr)
	m.Put("toAddress", email.ToAddress())
	m.Put("subject", email.Subject())
	m.Put("content", email.Body())
	m.Put("mailFormat", "plaintext")

	z.makeAPICallJson(z.AccountId(), "messages")

}
