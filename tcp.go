package gasp

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func ValidateSocket(socket string) error {
	if strings.TrimSpace(socket) == "" {
		return errors.New("socket is empty")
	}
	regEx, err := regexp.Compile("((([a-zA-Z\\-]+\\.)?([a-zA-Z\\-]+\\.)?([a-zA-Z\\-]+\\.)([a-zA-Z\\-]+))|((?:\\d{1,3}\\.){3}\\d{1,3})|(localhost)):\\d{1,5}")
	if err != nil {
		return fmt.Errorf("failed to compile regex: %v", err)
	}
	match := regEx.MatchString(socket)
	if !match {
		return fmt.Errorf("socket '%s' is invalid, should be in format <fqdn>:<port> or <ip_address>:<port>", socket)
	}
	return nil
}
