package commands

import "regexp"

var messageLinkRegex = regexp.MustCompile(`^https?:\/\/discord(?:app)?\.com\/channels\/\d+\/(\d+)\/(\d+)$`)
