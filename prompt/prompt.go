package prompt

import (
	"fmt"
	"sort"
	"strings"
	templates "text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/StiviiK/azctx/azurecli"
	"github.com/StiviiK/azctx/utils"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"golang.org/x/term"
)

// BuildPrompt builds a prompt for the user to select a subscription
func BuildPrompt(subscriptions utils.ComparableNamedSlice[azurecli.Subscription]) promptui.Select {
	// Get the terminal dimensions
	terminalWidth, terminalHeigth, err := term.GetSize(0)
	if err != nil {
		terminalWidth = 100 // Default width
		terminalHeigth = 20 // Default height
	}

	// Fetch the correct template
	tpl, shortPrompt := template(terminalWidth)

	// Sort the subscriptions by name
	sort.Sort(subscriptions)

	// Build the prompt
	subscriptionNames := utils.StringSlice(subscriptions.Names())
	maxSubscriptionsLength := subscriptionNames.LongestLength()
	maxTenantsLength := tenantNames(subscriptions, shortPrompt).LongestLength()

	return promptui.Select{
		Items: subscriptions,
		Templates: &promptui.SelectTemplates{
			Label:    fmt.Sprintf(tpl.Label, maxSubscriptionsLength, maxTenantsLength),
			Inactive: builItemTemplate(tpl.Inactive, maxSubscriptionsLength, maxTenantsLength, ""),
			Active:   builItemTemplate(tpl.Active, maxSubscriptionsLength, maxTenantsLength, "bold"),
			FuncMap:  newTemplateFuncMap(),
		},
		HideSelected: true,
		Searcher: func(input string, index int) bool {
			return fuzzy.MatchNormalized(strings.ToLower(input), strings.ToLower(subscriptionNames[index]))
		},
		Size:   utils.Min(len(subscriptions), utils.Max(utils.Min(terminalHeigth-3, 10), 1)),
		Stdout: utils.NoBellStdout,
	}
}

// buildItemTemplate builds the item template
func builItemTemplate(template string, maxSubscriptionsLength, maxTenantsLength int, additionalStyle string) string {
	return fmt.Sprintf(template, maxSubscriptionsLength, maxTenantsLength, additionalStyle)
}

// newTemplateFuncMap builds the template function map
func newTemplateFuncMap() templates.FuncMap {
	ret := sprig.TxtFuncMap()
	ret["green"] = promptui.Styler(promptui.FGGreen)
	ret["cyan"] = promptui.Styler(promptui.FGCyan)
	ret["bold"] = promptui.Styler(promptui.FGBold)
	ret["faint"] = promptui.Styler(promptui.FGFaint)
	return ret
}

// tenantNames returns the tenant names of the given subscriptions
func tenantNames(subscriptions []azurecli.Subscription, shortPrompt bool) utils.StringSlice {
	var tenantNames []string
	for _, subscription := range subscriptions {
		if !shortPrompt {
			tenantNames = append(tenantNames, fmt.Sprintf("%s (%s)", subscription.TenantName, subscription.Tenant))
		} else {
			tenantNames = append(tenantNames, subscription.TenantName)
		}
	}

	return tenantNames
}
