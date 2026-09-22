package scalingo

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	goscalingo "github.com/Scalingo/go-scalingo/v11"
	scalingohttp "github.com/Scalingo/go-scalingo/v11/http"
)

func TestDatabaseFeatureChanges(t *testing.T) {
	tests := map[string]struct {
		currentFeatures    []goscalingo.DatabaseFeature
		configuredFeatures []any
		wantAdded          []string
		wantRemoved        []string
	}{
		"does not enable an already enabled default feature": {
			currentFeatures: []goscalingo.DatabaseFeature{
				{Name: "redis-rdb"},
			},
			configuredFeatures: []any{"redis-aof", "redis-rdb", "force-ssl"},
			wantAdded:          []string{"redis-aof", "force-ssl"},
			wantRemoved:        []string{},
		},
		"removes a default feature absent from the configuration": {
			currentFeatures: []goscalingo.DatabaseFeature{
				{Name: "redis-rdb"},
			},
			configuredFeatures: []any{"redis-aof", "force-ssl"},
			wantAdded:          []string{"redis-aof", "force-ssl"},
			wantRemoved:        []string{"redis-rdb"},
		},
		"makes no changes when features match": {
			currentFeatures: []goscalingo.DatabaseFeature{
				{Name: "redis-aof"},
				{Name: "redis-rdb"},
				{Name: "force-ssl"},
			},
			configuredFeatures: []any{"redis-aof", "redis-rdb", "force-ssl"},
			wantAdded:          []string{},
			wantRemoved:        []string{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			added, removed, err := databaseFeatureChanges(test.currentFeatures, test.configuredFeatures)
			if err != nil {
				t.Fatalf("databaseFeatureChanges() error = %v", err)
			}
			if !reflect.DeepEqual(added, test.wantAdded) {
				t.Errorf("databaseFeatureChanges() added = %v, want %v", added, test.wantAdded)
			}
			if !reflect.DeepEqual(removed, test.wantRemoved) {
				t.Errorf("databaseFeatureChanges() removed = %v, want %v", removed, test.wantRemoved)
			}
		})
	}
}

func TestDatabaseFeatureChangesRejectsUnexpectedFeatureType(t *testing.T) {
	_, _, err := databaseFeatureChanges(nil, []any{42})
	if err == nil {
		t.Fatal("databaseFeatureChanges() error = nil, want an error")
	}
}

func TestDatabaseFeatureAlreadySetup(t *testing.T) {
	badRequestError := scalingohttp.BadRequestError{
		ErrMessage: "feature 'redis-rdb' is already setup on this database",
	}
	err := &scalingohttp.RequestFailedError{
		Code:     400,
		APIError: fmt.Errorf("enable feature: %w", badRequestError),
	}

	if !databaseFeatureAlreadySetup(err, "redis-rdb") {
		t.Error("databaseFeatureAlreadySetup() = false, want true")
	}
	if databaseFeatureAlreadySetup(err, "redis-aof") {
		t.Error("databaseFeatureAlreadySetup() = true for a different feature, want false")
	}
	if databaseFeatureAlreadySetup(errors.New("feature 'redis-rdb' is already setup on this database"), "redis-rdb") {
		t.Error("databaseFeatureAlreadySetup() = true for an untyped error, want false")
	}
}
