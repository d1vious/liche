package main

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/docopt/docopt-go"
)

const defaultConcurrency = maxOpenFiles / 2

const usage = `Link checker for Markdown and HTML

Usage:
	liche [-c <num-requests>] [-d <directory>] [-r] [-t <timeout>] [-x <regex>] [-v] <filenames>...

Options:
	-c, --concurrency <num-requests>  Set max number of concurrent HTTP requests. [default: %v]
	-d, --document-root <directory>  Set document root directory for absolute paths.
	-r, --recursive  Search Markdown and HTML files recursively
	-t, --timeout <timeout>  Set timeout for HTTP requests in seconds. Disabled by default.
	-x, --exclude <regex>  Regex of links to exclude from checking.
	-v, --verbose  Be verbose.`

type arguments struct {
	filenames    []string
	documentRoot string
	concurrency  int
	timeout      time.Duration
	recursive    bool
	exclude      *regexp.Regexp
	verbose      bool
}

func getArguments(argv []string) (arguments, error) {
	args, err := docopt.Parse(fmt.Sprintf(usage, defaultConcurrency), argv, true, "liche", true)

	if err != nil {
		return arguments{}, err
	}

	c, err := strconv.ParseInt(args["--concurrency"].(string), 10, 32)

	if err != nil {
		return arguments{}, err
	}

	if args["--document-root"] == nil {
		args["--document-root"] = ""
	}

	t := 0.0

	if args["--timeout"] != nil {
		t, err = strconv.ParseFloat(args["--timeout"].(string), 64)

		if err != nil {
			return arguments{}, err
		}
	}

	var exclude *regexp.Regexp
	if args["--exclude"] != nil {
		exclude, err = regexp.Compile(args["--exclude"].(string))
		if err != nil {
			return arguments{}, err
		}
	}

	return arguments{
		args["<filenames>"].([]string),
		args["--document-root"].(string),
		int(c),
		time.Duration(t) * time.Second,
		args["--recursive"].(bool),
		exclude,
		args["--verbose"].(bool),
	}, nil
}
