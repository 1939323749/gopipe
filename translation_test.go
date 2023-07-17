package main

import (
	"gopipe/utils"
	"testing"
)

func TestTranslation(t *testing.T) {
	_, err := utils.Translation("", "", "hello")
	if err != nil {
		t.Error(err)
	}
}
