/*
 * Signify app API
 *
 * API to access and configure the signify app
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package apiserver

// WidgetData - Data for a widget
type WidgetData struct {

	// The internal Id of widget data
	Id *int32 `json:"id,omitempty"`

	// Position of the element in widget type
	ElementSequence *int32 `json:"elementSequence,omitempty"`

	// The master asset id of this widget
	AssetId *int32 `json:"assetId,omitempty"`

	// individual config parameters depending on category
	Data *map[string]interface{} `json:"data,omitempty"`
}

// AssertWidgetDataRequired checks if the required fields are not zero-ed
func AssertWidgetDataRequired(obj WidgetData) error {
	return nil
}

// AssertWidgetDataConstraints checks if the values respects the defined constraints
func AssertWidgetDataConstraints(obj WidgetData) error {
	return nil
}
