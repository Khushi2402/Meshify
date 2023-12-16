package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"fmt"
	"context"
	"encoding/json"
	"time"
	"strings"
	"strconv"
	"regexp"
	

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"k8s.io/client-go/tools/clientcmd"
)

func getIngressIP() string {
	// Execute the kubectl command to update the istio-classic gateway [unable to locate the path]
	gatewayCmd := exec.Command("bash", "-c", "kubectl apply -f /Users/yashsharma/istio-1.17.2/samples/bookinfo/networking/bookinfo-gateway.yaml")
	_, err := gatewayCmd.Output()
	if err != nil {
		log.Printf("Failed to update istio-classic gateway: %s", err.Error())
		return ""
	}

	time.Sleep(10 * time.Second)



	// Execute the kubectl command to get the external IP address of the ingress controller
	ipcmd := exec.Command("bash", "-c", "kubectl get service -n istio-system istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}'")
	output, err := ipcmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to get Ingress IP address: %s", err.Error())
		return ""
	}

	ip:= string(output)

	return ip

	}

func extractTitle(html string) string {
	// Define a regular expression pattern to match the title tag
	titleRegex := regexp.MustCompile(`<title>(.*?)</title>`)
	match := titleRegex.FindStringSubmatch(html)

	if len(match) > 1 {
		title := match[1]
		title = strings.TrimSpace(title)
		title = strings.Trim(title, `"'`)

		return title
	}

	return ""
}

func runKubectlCommand(args ...string) string {
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Error executing kubectl command:", err.Error())
	}
	return string(out)
}

type LinkerdComponent struct {
	Name        string `json:"name"`
	ClusterIP   string `json:"clusterIP"`
}

type ServiceStatus struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}



func main() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

		// Set up JWT middleware for protected routes
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// Validate the token signing method and return the secret key
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid signing method")
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

		// Define a protected route that requires authentication(This isn't working)
		e.GET("/protected", func(c echo.Context) error {
			user := c.Get("user").(*jwt.Token)
			claims := user.Claims.(jwt.MapClaims)
			sub := claims["sub"].(string)
			return c.String(http.StatusOK, "You are authenticated as "+sub)
		}, echo.WrapMiddleware(jwtMiddleware.Handler))


			// Define a route for handling authentication with GitHub
	e.GET("/auth/github", func(c echo.Context) error {
		// Redirect the user to the GitHub authentication page
		err := godotenv.Load()
		if err != nil {
			fmt.Println("Error loading .env file")
			return nil
		}

		config := oauth2.Config{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			Endpoint:     github.Endpoint,
		}

		state := "T8f2Kbb3OaWkGNUYAqX8"
		c.SetCookie(&http.Cookie{
			Name:  "state",
			Value: state,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate random state string")
		}

		url := config.AuthCodeURL(state)
		return c.Redirect(http.StatusTemporaryRedirect, url)
	})

	e.GET("/auth/github/callback", func(c echo.Context) error {
		// Verify that the state parameter returned by GitHub matches the one in the cookie
		cookie, err := c.Cookie("state")
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Missing state cookie")
		}
		state := cookie.Value
		if c.QueryParam("state") != state {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid state parameter")
		}
		code := c.QueryParam("code")
		state = c.QueryParam("state")
		if state != "T8f2Kbb3OaWkGNUYAqX8" {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid state parameter")
		}
		// Exchange the authorization code for an access token
		config := oauth2.Config{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			Endpoint:     github.Endpoint,
		}
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to exchange authorization code")
		}
		// Get user info from the GitHub API
		client := config.Client(context.Background(), token)
		resp, err := client.Get("https://api.github.com/user")
		// Logging in the terminal
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to get user info")
		}
		defer resp.Body.Close()
		var user map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&user)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to decode user info")
		}
		// creating JWT token and giving baack to client
		jwtToken := jwt.New(jwt.SigningMethodHS256)
		claims := jwtToken.Claims.(jwt.MapClaims)
		claims["sub"] = user["login"].(string)
		claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
		tokenString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {

			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate JWT token")
		}

		// Set the token as a cookie
		cookie = &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Expires:  time.Now().Add(time.Hour * 72),
			HttpOnly: true,
			Secure:   true, // Set this to false if not using HTTPS
		}
		http.SetCookie(c.Response().Writer, cookie)

		// Redirect the user to the frontend page
		redirectURL := "http://localhost:3000/dashboard"
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)

	})

	

	if err := e.Start(":8080"); err != nil {
		panic(err)
	}
}

