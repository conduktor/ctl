package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOnlyPathWithSlashApi_UrlWithApi(t *testing.T) {
	inputUrl := "https://example.com/api/resource"
	expected := "/resource"
	result := getOnlyPathWithSlashApi(inputUrl)
	assert.Equal(t, expected, result)
}

func TestGetOnlyPathWithSlashApi_UrlWithoutApi(t *testing.T) {
	inputUrl := "https://example.com/resource"
	expected := "/resource"
	result := getOnlyPathWithSlashApi(inputUrl)
	assert.Equal(t, expected, result)
}

func TestGetOnlyPathWithSlashApi_PathWithoutApi(t *testing.T) {
	inputUrl := "https://example.com/resource"
	expected := "/resource"
	result := getOnlyPathWithSlashApi(inputUrl)
	assert.Equal(t, expected, result)
}

func TestGetOnlyPathWithSlashApi_PathWithApi(t *testing.T) {
	inputUrl := "api/resource"
	expected := "/resource"
	result := getOnlyPathWithSlashApi(inputUrl)
	assert.Equal(t, expected, result)
}
