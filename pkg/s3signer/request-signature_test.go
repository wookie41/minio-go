/*
 * Minio Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2015-2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package s3signer

import (
	"net/http"
	"strings"
	"testing"
	"io"
)

// Tests signature calculation.
func TestSignatureCalculation(t *testing.T) {
	req, err := http.NewRequest("GET", "https://s3.amazonaws.com", nil)
	if err != nil {
		t.Fatal("Error:", err)
	}
	req = SignV4(*req, "", "", "", "us-east-1")
	if req.Header.Get("Authorization") != "" {
		t.Fatal("Error: anonymous credentials should not have Authorization header.")
	}

	req = PreSignV4(*req, "", "", "", "us-east-1", 0)
	if strings.Contains(req.URL.RawQuery, "X-Amz-Signature") {
		t.Fatal("Error: anonymous credentials should not have Signature query resource.")
	}

	req = SignV2(*req, "", "")
	if req.Header.Get("Authorization") != "" {
		t.Fatal("Error: anonymous credentials should not have Authorization header.")
	}

	req = PreSignV2(*req, "", "", 0)
	if strings.Contains(req.URL.RawQuery, "Signature") {
		t.Fatal("Error: anonymous credentials should not have Signature query resource.")
	}

	req = SignV4(*req, "ACCESS-KEY", "SECRET-KEY", "", "us-east-1")
	if req.Header.Get("Authorization") == "" {
		t.Fatal("Error: normal credentials should have Authorization header.")
	}

	req = PreSignV4(*req, "ACCESS-KEY", "SECRET-KEY", "", "us-east-1", 0)
	if !strings.Contains(req.URL.RawQuery, "X-Amz-Signature") {
		t.Fatal("Error: normal credentials should have Signature query resource.")
	}

	req = SignV2(*req, "ACCESS-KEY", "SECRET-KEY")
	if req.Header.Get("Authorization") == "" {
		t.Fatal("Error: normal credentials should have Authorization header.")
	}

	req = PreSignV2(*req, "ACCESS-KEY", "SECRET-KEY", 0)
	if !strings.Contains(req.URL.RawQuery, "Signature") {
		t.Fatal("Error: normal credentials should not have Signature query resource.")
	}
}

func TestRequestHost(t *testing.T) {
	req, _ := buildRequest("dynamodb", "us-east-1", "{}")
	req.URL.RawQuery = "Foo=z&Foo=o&Foo=m&Foo=a"
	req.Host = "myhost"
	canonicalHeaders := getCanonicalHeaders(*req, v4IgnoredHeaders)

	if !strings.Contains(canonicalHeaders, "host:"+req.Host) {
		t.Errorf("canonical host header invalid")
	}
}

func buildRequest(serviceName, region, body string) (*http.Request, io.ReadSeeker) {
	endpoint := "https://" + serviceName + "." + region + ".amazonaws.com"
	reader := strings.NewReader(body)
	req, _ := http.NewRequest("POST", endpoint, reader)
	req.URL.Opaque = "//example.org/bucket/key-._~,!@#$%^&*()"
	req.Header.Add("X-Amz-Target", "prefix.Operation")
	req.Header.Add("Content-Type", "application/x-amz-json-1.0")
	req.Header.Add("Content-Length", string(len(body)))
	req.Header.Add("X-Amz-Meta-Other-Header", "some-value=!@#$%^&* (+)")
	req.Header.Add("X-Amz-Meta-Other-Header_With_Underscore", "some-value=!@#$%^&* (+)")
	req.Header.Add("X-amz-Meta-Other-Header_With_Underscore", "some-value=!@#$%^&* (+)")
	return req, reader
}