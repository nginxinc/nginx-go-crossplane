/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nginxinc/nginx-go-crossplane/internal/generator"
)

type filterFlag struct {
	filter map[string]struct{}
}

func (f *filterFlag) Set(value string) error {
	if f.filter == nil {
		f.filter = make(map[string]struct{})
	}
	f.filter[value] = struct{}{}
	return nil
}

func (f *filterFlag) String() string {
	return fmt.Sprintf("%v", *f)
}

type overrideItem struct {
	directive string
	masks     []generator.Mask
}

func (item *overrideItem) UnmarshalText(text []byte) error {
	rawOverride := string(text)

	// rawStr should follow the format: directive:bitmask00|bitmask01|...,bitmask10|bitmask11|...
	directive, definition, found := strings.Cut(rawOverride, ":")
	if !found {
		return errors.New("colon not found")
	}
	directive = strings.TrimSpace(directive)

	item.directive = directive
	if directive == "" {
		return errors.New("directive name is empty")
	}

	definition = strings.TrimSpace(definition)
	if definition == "" {
		return errors.New("directive definition is empty")
	}

	for _, varNamesStr := range strings.Split(definition, ",") {
		varNamesList := strings.Split(varNamesStr, "|")
		varNamesNum := len(varNamesList)
		directiveMask := make(generator.Mask, varNamesNum)

		for idx, varName := range varNamesList {
			trimmedName := strings.TrimSpace(varName)
			if trimmedName == "" {
				return errors.New("one directive bitmask is empty, check if there are unnecessary |")
			}

			directiveMask[idx] = trimmedName
		}
		item.masks = append(item.masks, directiveMask)
	}

	return nil
}

type override map[string][]generator.Mask

func (ov *override) String() string {
	if ov == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", *ov)
}

func (ov *override) Set(value string) error {
	if *ov == nil {
		*ov = override{}
	}
	var item overrideItem
	err := item.UnmarshalText([]byte(value))
	if err != nil {
		return fmt.Errorf("invalid override %s:%w", value, err)
	}

	(*ov)[item.directive] = item.masks
	return nil
}
