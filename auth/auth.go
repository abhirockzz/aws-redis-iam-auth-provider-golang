package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
)

const (
	// "The IAM authentication token is valid for 15 minutes"
	// https://docs.aws.amazon.com/memorydb/latest/devguide/auth-iam.html#auth-iam-limits
	tokenValiditySeconds = 900

	connectAction = "connect"

	// If the request has no payload you should use the hex encoded SHA-256 of an empty string as the payloadHash value.
	hexEncodedSHA256EmptyString = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

type AuthTokenGenerator struct {
	serviceName string
	region      string
	req         *http.Request

	credentials aws.Credentials
	signer      *v4.Signer
}

func New(serviceName, clusterName, userName, region string) (AuthTokenGenerator, error) {

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))

	if err != nil {
		return AuthTokenGenerator{}, err
	}

	credentials, err := cfg.Credentials.Retrieve(ctx)

	if err != nil {
		return AuthTokenGenerator{}, err
	}

	fmt.Println("credentials source -", credentials.Source)

	if credentials.AccessKeyID == "" || credentials.SecretAccessKey == "" {
		return AuthTokenGenerator{}, err
	}

	queryParams := url.Values{
		"Action":        {connectAction},
		"User":          {userName},
		"X-Amz-Expires": {strconv.FormatInt(int64(tokenValiditySeconds), 10)},
	}

	authURL := url.URL{
		Host:     clusterName,
		Scheme:   "http",
		Path:     "/",
		RawQuery: queryParams.Encode(),
	}

	req, err := http.NewRequest(http.MethodGet, authURL.String(), nil)

	if err != nil {
		return AuthTokenGenerator{}, err
	}

	return AuthTokenGenerator{
		serviceName: serviceName,
		region:      region,
		req:         req,
		credentials: credentials,
		signer:      v4.NewSigner(),
	}, nil
}

func (atg AuthTokenGenerator) Generate() (string, error) {

	signedURL, _, err := atg.signer.PresignHTTP(
		context.Background(),
		atg.credentials,
		atg.req,
		hexEncodedSHA256EmptyString,
		atg.serviceName,
		atg.region,
		time.Now().UTC(),
	)

	if err != nil {
		return "", fmt.Errorf("request signing failed - %w", err)
	}

	signedURL = strings.Replace(signedURL, "http://", "", 1)

	return signedURL, nil
}
