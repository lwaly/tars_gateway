package util

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lwaly/tars_gateway/common"
)

type Claims struct {
	Uid  uint64 `json:"uid"`
	Name string `json:"name"`
	jwt.StandardClaims
}

func ToMd5(uid uint64, expireToken int64, secret string) (decode []byte) {
	md5Ctx := md5.New()
	encode := fmt.Sprintf("%s%d%d", secret, uid, expireToken)
	md5Ctx.Write([]byte(encode))
	cipherStr := md5Ctx.Sum([]byte(secret))
	return []byte(base64.StdEncoding.EncodeToString(cipherStr))
}

func CreateToken(id uint64, second int64, secret, name string) string {
	expireToken := time.Now().Add(time.Second * time.Duration(second)).Unix()
	claims := Claims{
		id,
		name,
		jwt.StandardClaims{
			ExpiresAt: expireToken,
		},
	}

	// Create the token using your claims
	cToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signs the token with a secret.
	signedToken, _ := cToken.SignedString(ToMd5(claims.Uid, claims.ExpiresAt, secret))
	return signedToken
}

func TokenAuth(signedToken, secret string) (claims *Claims, err error) {
	s := strings.Split(signedToken, ".")
	if 3 != len(s) {
		return nil, errors.New(common.ErrNoUser)
	}

	claims = &Claims{}
	var token *jwt.Token
	var b []byte
	if b, err = jwt.DecodeSegment(s[1]); nil != err {
		return nil, err
	} else {
		if err := json.Unmarshal(b, claims); nil != err {
			return nil, err
		}
	}

	token, err = jwt.ParseWithClaims(signedToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return ToMd5(claims.Uid, claims.ExpiresAt, secret), nil
	})

	if nil != err {
		return nil, err
	}

	var ok bool
	if claims, ok = token.Claims.(*Claims); ok && token.Valid {
		return
	}

	return nil, err
}
