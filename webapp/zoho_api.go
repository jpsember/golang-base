package webapp

import (
	. "github.com/jpsember/golang-base/base"
	. "github.com/jpsember/golang-base/webapp/gen/webapp_data"
	"io"
	"net/http"
	"strings"
	"time"
)

type ZohoStruct struct {
	config   ZohoConfigBuilder
	modified bool
	bodyMap  JSMap
}

type Zoho = *ZohoStruct

func SharedZoho() Zoho {
	if sharedZoho == nil {
		sharedZoho = &ZohoStruct{}
		sharedZoho.initConfig()
	}
	return sharedZoho
}

func (z Zoho) initConfig() {
	var config ZohoConfig
	config = ProjStructure.Zoho()
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
	pr := PrIf("zoho_api;AccessToken", true)
	c := z.config
	if c.AccessToken() == "" || c.AccessTokenExpiryMs() < time.Now().UnixMilli() {
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
	pr := PrIf("AccountDetails", true)
	if z.config.AccountId() == "" {
		mp := z.makeAPICall()
		CheckState(mp.GetList("data").Length() == 1)
		data := mp.GetList("data").Get(0).AsJSMap()
		accountId := data.GetString("accountId")
		z.editConfig().SetAccountId(accountId)
		z.flushConfig()
		pr("account id:", accountId)
	}
	return z.config.AccountId()
}

func (z Zoho) editConfig() ZohoConfigBuilder {
	z.modified = true
	return z.config
}

func (z Zoho) flushConfig() {
	pr := PrIf("flushConfig", true)
	if z.modified {
		z.modified = false
		f := z.cacheFile()
		f.WriteStringM(z.config.ToJson().AsJSMap().CompactString())
		pr("flushed:", INDENT, z.config)
	}
}

var sharedZoho Zoho

func (z Zoho) Folders() map[string]string {
	pr := PrIf("Folders", true)
	if len(z.config.FolderMap()) == 0 {
		mp := z.makeAPICall(z.AccountId(), "folders")
		pr("results:", INDENT, mp)

		jl := mp.GetList("data")
		var x = make(map[string]string)
		for _, m := range jl.AsMaps() {
			x[m.GetString("folderName")] = m.GetString("folderId")
		}
		z.editConfig().SetFolderMap(x)
		z.flushConfig()
		pr("parsed:", INDENT, z.config.FolderMap())
	}
	return z.config.FolderMap()
}

func (z Zoho) makeAPICall(args ...any) JSMap {
	pr := PrIf("makeAPICall", true)

	method := http.MethodGet

	b := z.bodyMap
	z.bodyMap = nil
	if b != nil && false {
		// Setting method=Post seems to fail
		method = http.MethodPost
	}
	pr("args:", args)
	pr("body:", b)

	url := "https://mail.zoho.com/api/accounts"
	for _, x := range args {
		str := ToString(x)
		url = url + "/" + str
	}
	pr("url:", url)

	var body io.Reader
	if b != nil {
		s := strings.Builder{}
		s.WriteString(b.CompactString())
		body = io.NopCloser(strings.NewReader(s.String()))
	}

	client := &http.Client{}
	req := CheckOkWith(http.NewRequest(method, url, body))
	req.Header.Set("Authorization", "Zoho-oauthtoken "+z.AccessToken())

	resp := CheckOkWith(client.Do(req))

	defer resp.Body.Close()
	responseBody := CheckOkWith(io.ReadAll(resp.Body))

	pr("resp.Status:", resp.Status)
	pr("resp.Body:", INDENT, string(responseBody))

	mp := JSMapFromStringM(string(responseBody))
	Pr(mp)
	if mp.GetMap("status").GetInt("code") != 200 {
		BadState("returned unexpected status:", INDENT, mp)
	}
	return mp
}

func (z Zoho) body() JSMap {
	if z.bodyMap == nil {
		z.bodyMap = NewJSMap()
	}
	return z.bodyMap
}

func (z Zoho) ReadInbox() JSMap {
	pr := PrIf("ReadInbox", true)

	z.body().Put("folderId", z.Folders()["Inbox"])
	mp := z.makeAPICall(z.AccountId(), "messages", "view")
	Todo("How to put things in the request body?")

	pr("results:", INDENT, mp)
	return mp
}
