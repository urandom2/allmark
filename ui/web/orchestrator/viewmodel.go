// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package orchestrator

import (
	"github.com/andreaskoch/allmark2/common/index"
	"github.com/andreaskoch/allmark2/common/paths"
	"github.com/andreaskoch/allmark2/common/route"
	"github.com/andreaskoch/allmark2/model"
	"github.com/andreaskoch/allmark2/services/conversion"
	"github.com/andreaskoch/allmark2/ui/web/view/viewmodel"
)

func NewViewModelOrchestrator(itemIndex *index.ItemIndex, converter conversion.Converter) ViewModelOrchestrator {
	return ViewModelOrchestrator{
		itemIndex: itemIndex,
		converter: converter,
	}
}

type ViewModelOrchestrator struct {
	itemIndex *index.ItemIndex
	converter conversion.Converter
}

func (orchestrator *ViewModelOrchestrator) GetViewModel(pathProvider paths.Pather, item *model.Item) viewmodel.Model {

	// convert content
	convertedContent, err := orchestrator.converter.Convert(pathProvider, item)
	if err != nil {
		return viewmodel.Model{}
	}

	// create a view model
	viewModel := viewmodel.Model{
		Base: getBaseModel(item),

		Content: convertedContent,

		Childs:               orchestrator.getChildModels(item.Route()),
		ToplevelNavigation:   orchestrator.getToplevelNavigation(),
		BreadcrumbNavigation: orchestrator.getBreadcrumbNavigation(item),
		TopDocs:              orchestrator.getTopDocuments(5, pathProvider, item.Route()),
	}

	return viewModel
}

func (orchestrator *ViewModelOrchestrator) getTopDocuments(numberOfTopDocuments int, pathProvider paths.Pather, route *route.Route) []*viewmodel.Model {

	childItems := orchestrator.itemIndex.GetAllChilds(route)

	// determine the candidates for the top-documents
	candidateModels := make([]*viewmodel.Model, 0)

	for _, childItem := range childItems {

		if childItem.IsVirtual() {

			// the child is virtual: get the top documents of the child
			candidateModels = append(candidateModels, orchestrator.getTopDocuments(numberOfTopDocuments, pathProvider, childItem.Route())...)

		} else {

			// create viewmodel and append to list
			childModel := orchestrator.GetViewModel(pathProvider, childItem)
			candidateModels = append(candidateModels, &childModel)

		}

	}

	// sort the candidates
	viewmodel.SortModelBy(sortModelsByDateAndRoute).Sort(candidateModels)

	// take the top models only
	if len(candidateModels) <= numberOfTopDocuments {
		return candidateModels
	}

	return candidateModels[:numberOfTopDocuments]
}

func (orchestrator *ViewModelOrchestrator) getChildModels(route *route.Route) []*viewmodel.Base {
	childModels := make([]*viewmodel.Base, 0)

	childItems := orchestrator.itemIndex.GetChilds(route)
	for _, childItem := range childItems {
		baseModel := getBaseModel(childItem)
		childModels = append(childModels, &baseModel)
	}

	return childModels
}

func (orchestrator *ViewModelOrchestrator) getToplevelNavigation() *viewmodel.ToplevelNavigation {
	root, err := route.NewFromRequest("")
	if err != nil {
		return nil
	}

	toplevelEntries := make([]*viewmodel.ToplevelEntry, 0)
	for _, child := range orchestrator.itemIndex.GetChilds(root) {

		toplevelEntries = append(toplevelEntries, &viewmodel.ToplevelEntry{
			Title: child.Title,
			Path:  "/" + child.Route().Value(),
		})

	}

	return &viewmodel.ToplevelNavigation{
		Entries: toplevelEntries,
	}
}

func (orchestrator *ViewModelOrchestrator) getBreadcrumbNavigation(item *model.Item) *viewmodel.BreadcrumbNavigation {

	// create a new bread crumb navigation
	navigation := &viewmodel.BreadcrumbNavigation{
		Entries: make([]*viewmodel.Breadcrumb, 0),
	}

	// abort if item or model is nil
	if item == nil {
		return navigation
	}

	// recurse if there is a parent
	if parent := orchestrator.itemIndex.GetParent(item.Route()); parent != nil {
		navigation.Entries = append(navigation.Entries, orchestrator.getBreadcrumbNavigation(parent).Entries...)
	}

	// append a new navigation entry and return it
	navigation.Entries = append(navigation.Entries, &viewmodel.Breadcrumb{
		Title: item.Title,
		Level: item.Route().Level(),
		Path:  "/" + item.Route().Value(),
	})

	return navigation
}