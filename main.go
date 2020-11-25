package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"crypto/tls"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/qiangxue/go-env"
	log "github.com/sirupsen/logrus"
	//"github.com/docker/distribution/manifest/schema2"
	"gopkg.in/yaml.v3"
)

const (
	probeRepo = "nicokahlert/grafzahl"

	lOptFlag     = iota
	lOptEnv
	lOptConf
	lOptConfPath = "grafzahl.yaml"
)

var (
	remaining = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "docker_hub_rate_limit_remaining",
			Help: "The remaining pulls for this account.",
		},
	)
	limit = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "docker_hub_rate_limit",
			Help: "The maximal pulls for this account.",
		},
	)
	images = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "last_image_update_time",
			Help: "The last time an image available in this repo was updated.",
		},
	)
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
		},
	}
)

// Exporter structure
type Exporter struct {
	Password  string `yaml:"password",env:"PASSWORD"`
	Username  string `yaml:"username",env:"USERNAME"`
	Token     string
	Limit     int
	Remaining int
	Images    []int
	Catalogue []*Repository

}

// TokenResp from Dockerhub
type TokenResp struct {
	Token string `json:"token"`
	Details string `json:"details"`
}

// CatalogueResponse with available images in repo
type CatalogueResponse struct {
	Repositories []string `json:"repositories"`
}

//TagResponse F
type TagResponse struct {
	Tags []Tag `json:"tags"`
}

//Tag F
type Tag string

//Repository F
type Repository struct {
	Name string
	Tags map[Tag]int64
}

// SadManifesto F
type SadManifesto struct {
	History []struct {
		V1Compatibility string `json:"v1Compatibility"`
	} `json:"history"`
}

// GetToken from Dockerhub
func (e *Exporter) GetToken() error {
	log.Info("Asking for Token")
	tokenRequest, err := client.Get(fmt.Sprintf(
		"https://%s:%s@auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull",
		e.Username,
		e.Password,
		probeRepo,
	))
	if err != nil {
		log.Warn("Failed to connect to repository")
		return err
	}
	defer tokenRequest.Body.Close()
	tokenBytes, err := ioutil.ReadAll(tokenRequest.Body)
	if err != nil {
		log.Warn("Could not read body")
		return err
	}
	var tokenResp TokenResp
	err = json.Unmarshal(tokenBytes, &tokenResp)
	if err != nil {
		log.Warn("No valid json")
		return err
	}
	if tokenResp.Token == "" {
		log.Warn(tokenResp.Details)
		log.Fatal("No token!!! >:(")
	}
	e.Token = tokenResp.Token
	log.Info("Got Token :D")
	return nil
}

// Request your Limits and Remaining
func (e *Exporter) Request() error {
	req, err := http.NewRequest("HEAD", fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/latest", probeRepo), nil)
	if err != nil {
		log.Warn("Could not connect")
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", e.Token))
	res, err := client.Do(req)
	if err != nil {
		log.Warn("Failed to build request", err)
		return nil
	}
	rawLimit := res.Header.Get("RateLimit-Limit")
	if rawLimit == "" {
		log.Warn("Empty Header in Limits")
		return fmt.Errorf("Empty Headers in Limits %v", res.Body)
	}
	rawRemaining := res.Header.Get("RateLimit-Remaining")
	if rawRemaining == "" {
		log.Warn("Empty Header in Remaining", res.Body)
		return fmt.Errorf("Empty Headers in Remaining")
	}
	e.Limit, err = strconv.Atoi(strings.Split(rawLimit, ";")[0])
	e.Remaining, err = strconv.Atoi(strings.Split(rawRemaining, ";")[0])
	return err
}

// GetCatalogue checks the available Images
func (e *Exporter) GetCatalogue() error  {
	resp, err := http.Get("http://localhost:5000/v2/_catalog")
	if err != nil {
		log.Warn(err)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn(err)
		return err
	}
	var catalogueResponse CatalogueResponse
	if err := json.Unmarshal(body, &catalogueResponse); err != nil {
		log.Warn(err)
		return err
	}
	for _,c := range(catalogueResponse.Repositories){
		e.Catalogue = append(e.Catalogue, &Repository{
			Name: c,
			Tags: make(map[Tag]int64),
		})
	}
	return e.GetTags()
}

// GetTags F
func (e *Exporter) GetTags() error  {
	for _,r := range(e.Catalogue){
		resp, err := http.Get(fmt.Sprintf("http://localhost:5000/v2/%s/tags/list", r.Name ))
		if err != nil {
			log.Warn(err)
			return err
		}
		defer resp.Body.Close() //TODO auslagern!!!!!!
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warn(err)
			return err
		}
		var tagResponse TagResponse
		if err := json.Unmarshal(body, &tagResponse); err != nil {
			log.Warn(err)
			return err
		}
		for _,t := range(tagResponse.Tags){
			r.Tags[t]=0
		}
		r.GetTime()
	}
	return nil
}

func httpUnmarshall(url string, v interface{}) error {
	resp, err := http.Get(url)
		if err != nil {
			log.Warn(err)
			return err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warn(err)
			return err
		}
		return json.Unmarshal(body, v)
}

// GetTime F
func (r *Repository) GetTime() error {
	for tag := range(r.Tags) {
		var manifest SadManifesto
		if err := httpUnmarshall(
			fmt.Sprintf(
				"http://localhost:5000/v2/%s/manifests/%s",
				r.Name,
				tag,
			),
			&manifest,
		); err != nil {
			log.Warn(err)
			return err
		}
		compString := manifest.History[len(manifest.History)-1].V1Compatibility
		var comp map[string]interface{}
		if err := json.Unmarshal([]byte(compString), &comp); err != nil {
			log.Warn(err)
			return err
		}
		created, err := time.Parse("2006-01-02T15:04:05.000000000Z", comp["created"].(string))
		created_u64 := created.Unix
		if err != nil {
			log.Warn(err)
			return err
		}
		r.Tags[tag] = created_u64()
	}
	return nil
}

// RenewTokenEvery n hours
func (e *Exporter) RenewTokenEvery(n time.Duration) {
	go func() {
	  for {
		err := e.GetToken()
		if err != nil {
			log.Warn(err)
		}
		time.Sleep(n)
	  }
	}()
}

// RenewLimitValuesEvery n seconds
func (e *Exporter) RenewLimitValuesEvery(n time.Duration) {
	go func() {
	  for {
		err := e.Request()
		if err != nil {
			log.Warn(err)
		}
		limit.Set(
			float64(e.Limit))
		remaining.Set(
			float64(e.Remaining))
		log.Info(
			fmt.Sprintf(
				"Remaining Pulls: %d\tMax Pulls: %d",
				e.Remaining,
				e.Limit))
		time.Sleep(n)
	  }
	}()
}

// RenewImagesEvery renews images available in Registry every n seconds
func (e *Exporter)RenewImagesEvery(n time.Duration)  {
	go func () {
		for {
			err := e.GetCatalogue()
			if err != nil {
				log.Warn(err)
			}
			images.Set(3)
			time.Sleep(n)
			
		}
	}()
}

// Run the Exporter
func (e *Exporter) Run() {
	log.Info("Starting exporter on localhost:6969/metrics")
	prometheus.MustRegister(limit)
	prometheus.MustRegister(remaining)
	prometheus.MustRegister(images)
	e.RenewTokenEvery(2 * time.Hour)
	time.Sleep(3 * time.Second)
	e.RenewImagesEvery(3 * time.Second)
	e.RenewLimitValuesEvery(3 * time.Second)
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Endpoint ready")
	log.Fatal(http.ListenAndServe(":6969", nil))
}

// Init the Exporter via options
func (e *Exporter) Init(lOpt int) *Exporter {
	switch lOpt {
	case lOptFlag:
		flag.StringVar(&e.Username, "username", "", "Your docker-login username")
		flag.StringVar(&e.Password, "password", "", "Your docker-login password")
		flag.Parse()
		if e.Username == "" || e.Password == "" {
			log.Warn("No flags specified")
			return e.Init(lOptConf)
		}
		return e
	case lOptEnv:
		loader := env.New("GRAFZAHL_", nil)
		err := loader.Load(e)
		if err != nil {
			log.Fatal("Could not read from env", err)
			return nil
		}
		return e
	case lOptConf:
		data, err := ioutil.ReadFile(lOptConfPath)
		if err != nil {
			data, err = ioutil.ReadFile("/etc/" + lOptConfPath)
			if err != nil {
				log.Warn("Could not find config", err)
				return e.Init(lOptFlag)
			}
		}
		err = yaml.Unmarshal(data, e)
		if err != nil {
			log.Warn("Could not read YAML", err)
			return e.Init(lOptEnv)
		}
		return e
	default:
		return e.Init(lOptFlag)
	}
}

func main() {
	new(Exporter).Init(lOptFlag).Run()
}
