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
	"regexp"
	"strconv"


	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

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

	// API to retrieve the data of kubernetes cluster info
	e.GET("/api/get/cluster", func(c echo.Context) error {
		homedir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		// Load the kubeconfig file
		config, err := clientcmd.LoadFromFile(homedir + "/.kube/config")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		fmt.Println("Able to read kubeconfig")

		// Get the number of clusters
		numClusters := len(config.Clusters)

		// Create a slice to store the cluster information
		clusters := make([]map[string]string, 0)


		// Initialize a variable to store the current context
		var currentContext string

		// Get the current context from the kubeconfig
		if config.CurrentContext != "" {
			currentContext = config.CurrentContext
		}



		// Loop through the clusters and get their information
		for clusterName, cluster := range config.Clusters {
			// Get the cluster server URL
			serverURL := cluster.Server

			// Get the cluster certificate authority data
			caData := ""
			if cluster.CertificateAuthorityData != nil {
				caData = string(cluster.CertificateAuthorityData)
			}

				// Check if this is the current context
				isCurrentContext := false
				if clusterName == currentContext {
					isCurrentContext = true
				}



			// Add the cluster information to the slice
			clusters = append(clusters, map[string]string{
				"name":   clusterName,
				"server": serverURL,
				"caData": caData,
				"isactive":  strconv.FormatBool(isCurrentContext),
			})
		}

		data, err := json.Marshal(map[string]interface{}{
			"numClusters": numClusters,
			"clusters":    clusters,
			
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		c.Response().Header().Set("Content-Type", "application/json")
		return c.String(http.StatusOK, string(data))
})

		// API to get k8s information from the provided kubeconfig
		e.GET("/api/kube/info", func(c echo.Context) error {
			homedir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
			config, err := clientcmd.LoadFromFile(homedir + "/.kube/config")
			if err != nil {
				return err
			}
			restConfig, err := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
			if err != nil {
				return err
			}
			clientset, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				return err
			}
			ver, err := clientset.Discovery().ServerVersion()
			if err != nil {
				return err
			}
			svcList, err := clientset.CoreV1().Services("default").List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			podList, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			nodeList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			nodeNames := []string{}
			for _, node := range nodeList.Items {
				nodeNames = append(nodeNames, node.ObjectMeta.Name)
			}
			var svcTypes []string
			for _, svc := range svcList.Items {
				if svc.Spec.Type != "" {
					svcTypes = append(svcTypes, string(svc.Spec.Type))
				}
			}
			clusterInfo := map[string]interface{}{
				"serverVersion": ver,
				"numNodes":      len(nodeList.Items),
				"nodeNames":     nodeNames,
				"pods":          len(podList.Items),
				"svcTypes":      svcTypes,
			}
			data, err := json.Marshal(clusterInfo)
			if err != nil {
				return err
			}
			c.Response().Header().Set("Content-Type", "application/json")
			c.Response().WriteHeader(http.StatusOK)
			c.Response().Write(data)
			return nil
		})

		// API to get k8s workloads
		e.GET("/api/kube/workloads", func(c echo.Context) error {
			homedir, err := os.UserHomeDir()
			if err != nil {
			log.Fatal(err)
			}
		config, err := clientcmd.LoadFromFile(homedir + "/.kube/config")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		restConfig, err := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return err
		}

		clientset, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return err
		}

		ver, err := clientset.Discovery().ServerVersion()
		if err != nil {
			return err
		}
			nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			numNodes := len(nodes.Items)
			pods, err := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			services, err := clientset.CoreV1().Services("default").List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			podTemplates, err := clientset.CoreV1().PodTemplates("default").List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			replicationControllers, err := clientset.CoreV1().ReplicationControllers("default").List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			deployments, err := clientset.AppsV1().Deployments("default").List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			daemonSets, err := clientset.AppsV1().DaemonSets("default").List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			serviceNames := []string{}
	
			services, err = clientset.CoreV1().Services("").List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return err
			}
			for _, svc := range services.Items {
				serviceNames = append(serviceNames, svc.ObjectMeta.Name)
			}
			clusterInfo := map[string]interface{}{
				"serverVersion":             ver,
				"numNodes":                  numNodes,
				"numPods":                   len(pods.Items),
				"serviceNames":              serviceNames,
				"numServices":               len(services.Items),
				"numPodTemplates":           len(podTemplates.Items),
				"numReplicationControllers": len(replicationControllers.Items),
				"numDeployments":            len(deployments.Items),
				"numDaemonSets":             len(daemonSets.Items),
				
			}
			data, err := json.Marshal(clusterInfo)
			if err != nil {
				return err
			}
			c.Response().Header().Set("Content-Type", "application/json")
			c.Response().WriteHeader(http.StatusOK)
			c.Response().Write(data)
			return nil
	
		})

			// Endpoint to retirve the deployed and map them to the their ip addresses
	e.GET("/api/adapters", func(c echo.Context) error {

		// Load the Kubernetes configuration from the default location or from a specified file
		config, err := rest.InClusterConfig()
		if err != nil {
			config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"message": fmt.Sprintf("Failed to load Kubernetes configuration: %v", err),
				})
			}
		}
	
		// Create a new Kubernetes clientset using the loaded configuration
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Sprintf("Failed to create Kubernetes clientset: %v", err),
			})
		}

		selectedAdapter := c.QueryParam("adapter")
	
		// Get the list of all pods in the default namespace with the "app=adapter" label
		podList, err := clientset.CoreV1().Pods("istio-system").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Sprintf("Failed to get list of pods: %v", err),
			})
		}
	
		// Create a map of adapter name to IP address
		adapters := make(map[string]string)
		for _, pod := range podList.Items {
			adapters[pod.Name] = pod.Status.PodIP
		}

			// If the "adapter" query parameter is specified, return only that adapter
			if selectedAdapter != "" {
				ip, exists := adapters[selectedAdapter]
				if !exists {
					return c.JSON(http.StatusNotFound, map[string]string{
						"message": fmt.Sprintf("Adapter not found: %s", selectedAdapter),
					})
				}
				adapters = map[string]string{selectedAdapter: ip}
			}
	
		return c.JSON(http.StatusOK, adapters)
})

		// Deploying bookinfo.yaml application in default namespace
		e.POST("/api/deploy/bookinfo", func(c echo.Context) error {
	
			// Execute the kubectl command to apply the Bookinfo YAML
			deployCmd := exec.Command("kubectl", "apply", "-f", "/Users/yashsharma/istio-1.17.2/samples/bookinfo/platform/kube/bookinfo.yaml")
			deployCmd.Stdout = os.Stdout
			deployCmd.Stderr = os.Stderr
			err := deployCmd.Run()
			if err != nil {
				log.Printf("Failed to deploy Bookinfo application: %s", err.Error())
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"message": "Failed to deploy Bookinfo application",
				})
			}
			
				// Execute the kubectl command to retrieve the title of the product page
				titleCmd := exec.Command("bash", "-c", "kubectl exec $(kubectl get pod -l app=ratings -o jsonpath='{.items[0].metadata.name}') -c ratings -- curl -sS productpage:9080/productpage | grep -o '<title>.*</title>'")
				output, err := titleCmd.CombinedOutput()
				if err != nil {
					log.Printf("Failed to retrieve product page title: %s", err.Error())
					return c.JSON(http.StatusInternalServerError, map[string]string{
						"message": "Failed to retrieve product page title, Wait for the application to be deployed",
					})
				}
	
				title := extractTitle(string(output))
				if title == "" {
					log.Println("Failed to extract product page title")
					return c.JSON(http.StatusInternalServerError, map[string]string{
						"message": "Failed to extract product page title",
					})
				}

			ip := getIngressIP()
			if ip == "" {
				log.Println("Make sure your cluster has an Ingress controller running and  it has an external IP address")
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"message": "make sure your cluster has an Ingress controller running and  it has an external IP address",
				})
				}	
		
			log.Printf("Bookinfo application deployed successfully. Ingress IP: %s", ip)
			return c.JSON(http.StatusOK, map[string]string{
				"title": title,
				"message":    "Bookinfo application deployed successfully",
				"ingress_ip": ip,
	
			})
		})

								//  Metrics and monitoring

		e.GET("/api/prometheusgrafana/health/status", func(c echo.Context) error {
		// Check Prometheus status
		prometheusIP := runKubectlCommand("get", "services", "-n", "istio-system", "-l", "app=prometheus", "-o", "jsonpath='{.items[*].spec.clusterIP}'")

		prometheus := ServiceStatus{
			Name:    "Prometheus",
			Address: strings.Trim(prometheusIP, "'"),
		}

		// Check Grafana status
		grafanaIP := runKubectlCommand("get", "services", "-n", "istio-system", "-l", "app=grafana", "-o", "jsonpath='{.items[*].spec.clusterIP}'")

		grafana := ServiceStatus{
			Name:    "Grafana",
			Address: strings.Trim(grafanaIP, "'"),
		}

		status := []ServiceStatus{prometheus, grafana}

		return c.JSON(http.StatusOK, status)
	})

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if he, ok := err.(*echo.HTTPError); ok {
			if he.Code == http.StatusNotFound {
				log.Println("Error 404: Not Found")
			}
		}
		c.JSON(http.StatusInternalServerError, err.Error())
	}


	e.GET("/api/dashboard/prometheus", func(c echo.Context) error {
		// Execute the 'istioctl dashboard prometheus' command
		cmd := exec.Command("istioctl", "dashboard", "prometheus")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	
		err := cmd.Run()
		if err != nil {
			log.Println(err)
			return c.String(http.StatusInternalServerError, "Error executing command")
		}
	
		// Check the command's exit status code
		if cmd.ProcessState.ExitCode() == 0 {
			// If the exit status is 0 (success), return the URL and open it in a new tab
			return c.Redirect(http.StatusOK, "http://localhost:9090")
		}
	
		return c.String(http.StatusInternalServerError, "Command executed, but encountered an error")
	})
	

	if err := e.Start(":8080"); err != nil {
		panic(err)
	}
}

