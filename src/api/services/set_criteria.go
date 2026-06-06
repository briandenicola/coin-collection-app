package services

import (
	"fmt"
	"slices"
)

var allowedCriteriaFields = []string{
	"material", "category", "denomination", "ruler", "era", "mint", "grade",
	"currentValue", "purchasePrice", "purchaseDate", "createdAt", "isWishlist",
	"isSold", "isPrivate",
}

var allowedCriteriaOps = []string{
	"eq", "neq", "contains", "startsWith", "in", "between", "gte", "lte", "isNull", "isNotNull",
}

// ValidateSmartCriteria validates the restricted smart-set criteria tree.
func ValidateSmartCriteria(criteria map[string]interface{}) error {
	if len(criteria) == 0 {
		return fmt.Errorf("smart criteria is required")
	}
	return validateCriteriaNode(criteria)
}

func validateCriteriaNode(node map[string]interface{}) error {
	if op, ok := node["operator"].(string); ok {
		if op != "and" && op != "or" {
			return fmt.Errorf("criteria operator must be and or or")
		}
		rules, ok := node["rules"].([]interface{})
		if !ok || len(rules) == 0 {
			return fmt.Errorf("criteria groups require rules")
		}
		for _, raw := range rules {
			child, ok := raw.(map[string]interface{})
			if !ok {
				return fmt.Errorf("criteria rule must be an object")
			}
			if err := validateCriteriaNode(child); err != nil {
				return err
			}
		}
		return nil
	}

	field, _ := node["field"].(string)
	ruleOp, _ := node["op"].(string)
	if !slices.Contains(allowedCriteriaFields, field) {
		return fmt.Errorf("criteria field is not allowed")
	}
	if !slices.Contains(allowedCriteriaOps, ruleOp) {
		return fmt.Errorf("criteria operator is not allowed")
	}
	return nil
}
