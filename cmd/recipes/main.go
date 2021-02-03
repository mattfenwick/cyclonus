package main

import (
	"fmt"
	"github.com/mattfenwick/cyclonus/pkg/explainer"
	"github.com/mattfenwick/cyclonus/pkg/matcher"
	"github.com/mattfenwick/cyclonus/pkg/recipes"
	"github.com/mattfenwick/cyclonus/pkg/utils"
)

func main() {
	for _, recipe := range recipes.AllRecipes {
		result, err := recipe.RunProbe()
		utils.DoOrDie(err)

		explainer.TableExplainer(matcher.BuildNetworkPolicies(recipe.Policies())).Render()

		fmt.Printf("resources:\n%s\n", recipe.Resources.Table())

		fmt.Printf("ingress:\n%s\n", result.Ingress.Table())

		fmt.Printf("egress:\n%s\n", result.Egress.Table())

		fmt.Printf("combined:\n%s\n", result.Combined.Table())

		fmt.Printf("\n\n\n")
	}
}
