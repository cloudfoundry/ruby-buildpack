package integration_test

import (
	"fmt"
	"regexp"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

type haveUnchangedAppdirMatcher struct {
	before, after                 string
	checksumBefore, checksumAfter string
}

func HaveUnchangedAppdir(before, after string) types.GomegaMatcher {
	return &haveUnchangedAppdirMatcher{before: before, after: after}
}

func (matcher *haveUnchangedAppdirMatcher) Match(actual interface{}) (success bool, err error) {
	app, ok := actual.(*cutlass.App)
	if !ok {
		return false, fmt.Errorf("HaveUnchangedAppdir matcher requires a cutlass.App.  Got:\n%s", format.Object(actual, 1))
	}

	re := regexp.MustCompile(fmt.Sprintf("%s: ([0-9a-f]+)", matcher.before))
	matches := re.FindAllStringSubmatch(app.Stdout.String(), -1)
	if len(matches) < 1 || len(matches[0]) < 2 {
		return false, nil
	}
	matcher.checksumBefore = matches[0][1]

	re = regexp.MustCompile(fmt.Sprintf("%s: ([0-9a-f]+)", matcher.after))
	matches = re.FindAllStringSubmatch(app.Stdout.String(), -1)
	if len(matches) < 1 || len(matches[0]) < 2 {
		return false, nil
	}
	matcher.checksumAfter = matches[0][1]

	return matcher.checksumBefore != "" && matcher.checksumBefore == matcher.checksumAfter, nil
}

func (matcher *haveUnchangedAppdirMatcher) FailureMessage(actual interface{}) (message string) {
	app := actual.(*cutlass.App)
	if matcher.checksumBefore == "" {
		return format.Message(app.Stdout.String(), "to contain substring", matcher.before)
	} else if matcher.checksumAfter == "" {
		return format.Message(app.Stdout.String(), "to contain substring", matcher.after)
	}
	return format.Message(matcher.checksumBefore, "to be the same checksum as", matcher.checksumAfter)
}

func (matcher *haveUnchangedAppdirMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	app := actual.(*cutlass.App)
	if matcher.checksumBefore == "" {
		return format.Message(app.Stdout.String(), "to contain substring", matcher.before)
	} else if matcher.checksumAfter == "" {
		return format.Message(app.Stdout.String(), "to contain substring", matcher.after)
	}
	return format.Message(matcher.checksumBefore, "to be a different checksum to", matcher.checksumAfter)
}
