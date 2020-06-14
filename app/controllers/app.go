package controllers

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/revel/revel"
	"html"
	"regexp"
	"strings"
	"time"
)

// For HMAC signing method, the key can be any []byte. It is recommended to generate
// a key using crypto/rand or something equivalent. You need the same key for signing
// and validating.
var hmacSampleSecret []byte

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

/*

Creds are:

sid / ThisIsLongSecurePassword

*/

func (c App) ClientSide(hash string) revel.Result {
	logged_in := false
	if hash == "e2b18481be9c7b210e3fa881d7484495" {
		logged_in = true
	}

	return c.Render(logged_in)
}

func (c App) Timing() revel.Result {
	return c.Render()
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (c App) Timing_Login(username, password string) revel.Result {
	users := make([]string, 4)
	users[0] = "zoe"
	users[1] = "joe"
	users[2] = "alex"
	users[3] = "sarah"

	if contains(users, strings.ToLower(username)) {
		time.Sleep(3000 * time.Millisecond)

	}
	c.Flash.Error("Login Failed")
	c.FlashParams()
	return c.Redirect(App.Timing)
}

/*

Auth0 vulnerability

https://auth0.com/docs/security/bulletins/cve-2019-7644

https://godoc.org/github.com/dgrijalva/jwt-go#example-Parse--ErrorChecking

*/

func getToken(token *jwt.Token) (interface{}, error) {
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}

	// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
	hmacSampleSecret = []byte("my_secret_key")
	return hmacSampleSecret, nil
}

func ParseJWT(tokenString string) (bool, string) {
	var success bool = false
	var message string = ""

	token, err := jwt.Parse(tokenString, getToken)

	if err == nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			message = fmt.Sprintf("Welcome %s (%s)", claims["user"], claims["level"])
			success = true
		}
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		// This is from <https://godoc.org/github.com/dgrijalva/jwt-go#pkg-constants>
		if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
			newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token.Claims)
			newTokenString, _ := newToken.SignedString(hmacSampleSecret)

			newParsedToken, _ := jwt.Parse(newTokenString, getToken)

			message = fmt.Sprintf("Invalid signature. Expected %s got %s", newParsedToken.Signature, token.Signature)
			//message = fmt.Sprintf("err: %s", err)
		} else if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			message = fmt.Sprintln("That's not even a token")
		} else {
			fmt.Sprintln("Couldn't handle this token:", err)
		}
	} else {
		message = fmt.Sprintf("There was an error parsing the token: %s", err.Error())
	}

	return success, message
}

func (c App) Auth1_Login(jwt string) revel.Result {
	success, message := ParseJWT(jwt)

	if success {
		c.Flash.Success(message)
	} else {
		c.Flash.Error(message)
	}
	c.FlashParams()

	return c.Redirect(App.Auth1)
}
func (c App) Auth1() revel.Result {
	/*
		// Create a new token object, specifying signing method and the claims
		// you would like it to contain.
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": "joe",
			"password": "2ac9cb7dc02b3c0083eb70898e549b63",
			"level":    "admin",
		})

		// Sign and get the complete encoded token as a string using the secret
		tokenString, _ := token.SignedString(hmacSampleSecret)
	*/

	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsZXZlbCI6InVzZXIiLCJ1c2VyIjoic2lkIn0.Hnpn5k6NtrXn8qvOuiSsFjXhAolQGn3TfmGBvA7EGTU"

	//var username string = c.Params.Form["username"][0]
	username := ""
	return c.Render(tokenString, username)
}

func (c App) Bypass() revel.Result {
	forwarded := c.Request.Header.Get("X-Forwarded-For")
	logged_in := false
	if strings.Contains(forwarded, "192.168.0.14") {
		logged_in = true
	}

	return c.Render(forwarded, logged_in)
}

/******************

Expired JWT Lab

******************/

func (c App) Expired_JWT_Login(username, password string) revel.Result {
	if username == "joe" && password == "Password1" {
		c.Flash.Success("Login Success")
	} else {
		c.Flash.Error("Login Failed")
	}
	c.FlashParams()

	return c.Redirect(App.Expired_JWT)
}
func (c App) Expired_JWT() revel.Result {
	/*
		// Create a new token object, specifying signing method and the claims
		// you would like it to contain.
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": "joe",
			"password": "2ac9cb7dc02b3c0083eb70898e549b63",
			"level":    "admin",
		})

		// Sign and get the complete encoded token as a string using the secret
		tokenString, _ := token.SignedString(hmacSampleSecret)
	*/

	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsZXZlbCI6ImFkbWluIiwicGFzc3dvcmQiOiIyYWM5Y2I3ZGMwMmIzYzAwODNlYjcwODk4ZTU0OWI2MyIsInVzZXJuYW1lIjoiam9lIn0.6j3NrK-0C7K8gmaWeB9CCyZuQKfvVEAl4KhitRN2p5k"

	//var username string = c.Params.Form["username"][0]
	username := ""
	return c.Render(tokenString, username)
}

/******************

Leaky JWT Lab

******************/

func (c App) Leaky_JWT_Login(username, password string) revel.Result {
	if username == "joe" && password == "Password1" {
		c.Flash.Success("Login Success")
	} else {
		c.Flash.Error("Login Failed")
	}
	c.FlashParams()

	return c.Redirect(App.Leaky_JWT)
}
func (c App) Leaky_JWT() revel.Result {
	/*
		// Create a new token object, specifying signing method and the claims
		// you would like it to contain.
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": "joe",
			"password": "2ac9cb7dc02b3c0083eb70898e549b63",
			"level":    "admin",
		})

		// Sign and get the complete encoded token as a string using the secret
		tokenString, _ := token.SignedString(hmacSampleSecret)
	*/

	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsZXZlbCI6ImFkbWluIiwicGFzc3dvcmQiOiIyYWM5Y2I3ZGMwMmIzYzAwODNlYjcwODk4ZTU0OWI2MyIsInVzZXJuYW1lIjoiam9lIn0.6j3NrK-0C7K8gmaWeB9CCyZuQKfvVEAl4KhitRN2p5k"

	//var username string = c.Params.Form["username"][0]
	username := ""
	return c.Render(tokenString, username)
}

/******************

User Agent Bypass

******************/

func (c App) UserAgent() revel.Result {
	ua := c.Request.Header.Get("User-Agent")

	app := ""
	comment := fmt.Sprintf("<!-- For debug, the user agent supplied is: %s -->", html.EscapeString(ua))
	if ua == "authlab desktop app" {
		app = "desktop"
	}

	return c.Render(app, comment)
}

func (c App) UserAgent_Ping() revel.Result {
	ua := c.Request.Header.Get("User-Agent")

	app := ""
	if ua == "authlab desktop app" {
		app = "desktop"
	}

	return c.Render(app)
}

/******************

JWT None

******************/

func ParseJWTNone(tokenString string) (bool, string) {
	var success bool = false
	var message string = ""
	var algorithm string = ""

	token, err := jwt.Parse(tokenString, getToken)
	// Token is one of these
	// https://godoc.org/github.com/dgrijalva/jwt-go#Token

	if err != nil {
		fmt.Printf("Error parsing token\n")
		return false, "Error parsing token"
	}

	// fmt.Printf("TokenString is: %s\n", tokenString)
	// fmt.Printf("Token is: %u\n", token)

	if token.Method != nil {
		fmt.Printf("Algorithm is: %s\n", token.Method.Alg())
		algorithm = token.Method.Alg()
	} else {
		fmt.Printf("Unknown algorithm passed\n")
		return false, "Unknown hashing algorithm"
	}
	// Check if algorithm is none, if so, return the data, if anything else, then parse as it should be

	return true, "aaa"

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		message = fmt.Sprintf("Welcome %s (%s)", claims["user"], claims["level"])
		success = true
	}
	/*
		} else if ve, ok := err.(*jwt.ValidationError); ok {
			// This is from <https://godoc.org/github.com/dgrijalva/jwt-go#pkg-constants>
			if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token.Claims)
				newTokenString, _ := newToken.SignedString(hmacSampleSecret)

				newParsedToken, _ := jwt.Parse(newTokenString, getToken)

				message = fmt.Sprintf("Invalid signature. Expected %s got %s", newParsedToken.Signature, token.Signature)
				//message = fmt.Sprintf("err: %s", err)
			} else if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				message = fmt.Sprintln("That's not even a token")
			} else {
				fmt.Sprintln("Couldn't handle this token:", err)
			}
		} else {
			message = fmt.Sprintf("There was an error parsing the token: %s", err.Error())
		}
	*/

	return success, message
}

func (c App) JWT_None_Check() revel.Result {
	data := make(map[string]interface{})

	bearer_header := c.Request.GetHttpHeader("Authorization")
	// fmt.Printf("headers %s\n", bearer_header)

	re := regexp.MustCompile("(?i)^bearer (.*)$")
	//re := regexp.MustCompile("(?i)bearer ([.0-9a-z=])$")
	matches := re.FindStringSubmatch(bearer_header)
	// fmt.Printf("Length of tokens: %d\n", len(matches))

	jwt := ""
	if len(matches) == 2 {
		fmt.Printf("Hit for token: %s\n", matches[1])
		jwt = matches[1]
	} else {
		fmt.Printf("Login failed\n")
		data["error"] = true
		data["stuff"] = "No token found"
		return c.RenderJSON(data)
	}

	success, message := ParseJWTNone(jwt)

	user := "robin"
	if success {
		fmt.Printf("Login success\n")
		data["error"] = false
		data["stuff"] = fmt.Sprintf("Logged in as %s", user)
	} else {
		fmt.Printf("Login failed\n")
		data["error"] = true
		data["stuff"] = message
	}
	return c.RenderJSON(data)
}

func (c App) JWT_None() revel.Result {
	return c.Render()
}
