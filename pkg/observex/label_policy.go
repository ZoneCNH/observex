package observex

import (
	"fmt"
	"strings"
)

// LabelPolicy defines which label names are allowed or denied for metric use.
type LabelPolicy struct {
	// Allowed is the set of explicitly permitted label names.
	// When non-nil and non-empty, only labels in this set are accepted.
	Allowed map[string]struct{}
	// Denied is the set of label names that must not be used.
	Denied map[string]struct{}
}

// DefaultDeniedLabels are labels that must not be used as metric keys
// because they are ambiguous or shadow well-known log fields.
var DefaultDeniedLabels = []string{"error", "err", "msg", "level"}

// NewDefaultLabelPolicy returns a LabelPolicy with the default denied list.
func NewDefaultLabelPolicy() LabelPolicy {
	denied := make(map[string]struct{}, len(DefaultDeniedLabels))
	for _, name := range DefaultDeniedLabels {
		denied[name] = struct{}{}
	}
	return LabelPolicy{Denied: denied}
}

// NewLabelPolicy returns a LabelPolicy with explicit allowed and denied sets.
func NewLabelPolicy(allowed []string, denied []string) LabelPolicy {
	p := LabelPolicy{}
	if len(allowed) > 0 {
		p.Allowed = make(map[string]struct{}, len(allowed))
		for _, name := range allowed {
			p.Allowed[name] = struct{}{}
		}
	}
	if len(denied) > 0 {
		p.Denied = make(map[string]struct{}, len(denied))
		for _, name := range denied {
			p.Denied[name] = struct{}{}
		}
	}
	return p
}

// ValidateLabel checks a single label name against this policy.
// It returns an error if the label violates naming rules or the deny list.
func (p LabelPolicy) ValidateLabel(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("label policy: label name must not be empty")
	}

	if !labelKeyRE.MatchString(name) {
		return fmt.Errorf("label policy: label %q must be lower snake case starting with a letter", name)
	}

	if _, denied := p.Denied[name]; denied {
		return fmt.Errorf("label policy: label %q is in the denied list", name)
	}

	if p.Allowed != nil {
		if _, allowed := p.Allowed[name]; !allowed {
			return fmt.Errorf("label policy: label %q is not in the allowed list", name)
		}
	}

	return nil
}

// ValidateLabels checks multiple label names against this policy.
// It returns all validation errors found, or nil if all names are valid.
func (p LabelPolicy) ValidateLabels(names []string) []error {
	var errs []error
	for _, name := range names {
		if err := p.ValidateLabel(name); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
