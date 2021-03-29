package verifier_mailru

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/skeris/identity/identity"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/mailru"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

var _ identity.OAuth2Verifier = new(Verifier)

type Verifier struct {
	oacfg *oauth2.Config
}

const app_id = ""
const secret_key = ""

func New(cfg Config) *Verifier {
	prov := &Verifier{
		oacfg: &oauth2.Config{
			RedirectURL:  cfg.RedirectURL,
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Scopes:       cfg.Scopes,
			Endpoint:     mailru.Endpoint,
		},
	}

	return prov
}

func (prov *Verifier) Info() identity.VerifierInfo {
	return identity.VerifierInfo{
		Name: "mailru",
	}
}

func (prov *Verifier) NormalizeIdentity(idn string) string {
	return idn
}

func (prov *Verifier) GetOAuth2URL(state string) string {
	return prov.oacfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (prov *Verifier) HandleOAuth2Callback(ctx context.Context, code string) (token *oauth2.Token, err error) {
	token, err = prov.oacfg.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func strMD5(str string) string {
	tstr := md5.Sum([]byte(str))
	return hex.EncodeToString(tstr[:])
}

func (prov *Verifier) GetOAuth2Identity(ctx context.Context, accessToken string) (iden *identity.IdentityData, verifierData *identity.VerifierData, err error) {
	u, err := url.Parse("http://www.appsmail.ru/platform/api")
	if err != nil {
		return nil, nil, err
	}
	query := url.Values{
		"method":      {"users.getInfo"},
		"app_id":      {app_id},
		"session_key": {accessToken},
		"secure":      {"1"},
		"format":      {"json"},
		"sig":         {strMD5("app_id=" + app_id + "method=friends.getsecure=1session_key=" + accessToken + secret_key)},
	}
	u.RawQuery = query.Encode()

	client := &http.Client{}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	var userInfo UserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, nil, err
	}
	if userInfo[0].Error.ErrorMsg != "" {
		return nil, nil, errors.New(userInfo[0].Error.ErrorMsg)
	}
	return &identity.IdentityData{}, &identity.VerifierData{
		Name: "mailru",
		AuthenticationData: nil,
		AdditionalData: identity.B{"mailru": []byte(data[:])},
	}, nil
}

type UserInfo []struct {
	Error struct {
		ErrorMsg   string `json:"error_msg"`
		ErrorToken string `json:"error_token"`
		Extended   string `json:"extended"`
		ErrorCode  int    `json:"error_code"`
	} `json:"error"`
	Uid          string `json:"uid"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Nick         string `json:"nick"`
	Email        string `json:"email"`
	Sex          int    `json:"sex"`
	Birthday     string `json:"birthday"`
	HasPic       int    `json:"has_pic"`
	PicBig       string `json:"pic_big"`
	Link         string `json:"link"`
	RefererType  string `json:"referer_type"`
	RefererId    string `json:"referer_id"`
	FriendsCount int    `json:"friends_count"`
	IsVerified   int    `json:"is_verified"`
	Vip          int    `json:"vip"`
	Location     struct {
		Country struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		} `json:"country"`
		City struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		} `json:"city"`

		Region struct {
			Name string `json:"name"`
			Id   string `json:"id"`
		} `json:"region"`
	} `json:"location"`
}
