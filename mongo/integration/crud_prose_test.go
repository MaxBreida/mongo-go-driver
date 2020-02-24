// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package integration

import (
	"errors"
	"testing"

	"github.com/MaxBreida/mongo-go-driver/bson"
	"github.com/MaxBreida/mongo-go-driver/internal/testutil/assert"
	"github.com/MaxBreida/mongo-go-driver/mongo"
	"github.com/MaxBreida/mongo-go-driver/mongo/integration/mtest"
	"github.com/MaxBreida/mongo-go-driver/mongo/options"
)

func TestCrudProse(t *testing.T) {
	clientOpts := options.Client().SetRetryWrites(false).SetWriteConcern(mtest.MajorityWc).
		SetReadConcern(mtest.MajorityRc)
	mtOpts := mtest.NewOptions().ClientOptions(clientOpts).MinServerVersion("4.0").Topologies(mtest.ReplicaSet).
		CreateClient(false)
	mt := mtest.New(t, mtOpts)
	defer mt.Close()

	label := "ExampleError"
	mt.Run("InsertMany errors with label", func(mt *mtest.T) {
		mt.SetFailPoint(mtest.FailPoint{
			ConfigureFailPoint: "failCommand",
			Mode: mtest.FailPointMode{
				Times: 1,
			},
			Data: mtest.FailPointData{
				FailCommands: []string{"insert"},
				WriteConcernError: &mtest.WriteConcernErrorData{
					Code:        100,
					ErrorLabels: &[]string{label},
				},
			},
		})

		_, err := mt.Coll.InsertMany(mtest.Background,
			[]interface{}{
				bson.D{
					{"a", 1},
				},
				bson.D{
					{"a", 2},
				},
			})
		assert.NotNil(mt, err, "expected non-nil error, got nil")

		we, ok := err.(mongo.BulkWriteException)
		assert.True(mt, ok, "expected mongo.BulkWriteException, got %T", err)
		assert.True(mt, we.HasErrorLabel(label), "expected error to have label: %v", label)
	})

	mt.Run("WriteException with label", func(mt *mtest.T) {
		mt.SetFailPoint(mtest.FailPoint{
			ConfigureFailPoint: "failCommand",
			Mode: mtest.FailPointMode{
				Times: 1,
			},
			Data: mtest.FailPointData{
				FailCommands: []string{"delete"},
				WriteConcernError: &mtest.WriteConcernErrorData{
					Code:        100,
					ErrorLabels: &[]string{label},
				},
			},
		})

		_, err := mt.Coll.DeleteMany(mtest.Background, bson.D{{"a", 1}})
		assert.NotNil(mt, err, "expected non-nil error, got nil")

		we, ok := err.(mongo.WriteException)
		assert.True(mt, ok, "expected mongo.WriteException, got %T", err)
		assert.True(mt, we.HasErrorLabel(label), "expected error to have label: %v", label)
	})

	mt.Run("BulkWriteException with label", func(mt *mtest.T) {
		mt.SetFailPoint(mtest.FailPoint{
			ConfigureFailPoint: "failCommand",
			Mode: mtest.FailPointMode{
				Times: 1,
			},
			Data: mtest.FailPointData{
				FailCommands: []string{"delete"},
				WriteConcernError: &mtest.WriteConcernErrorData{
					Code:        100,
					ErrorLabels: &[]string{label},
				},
			},
		})

		models := []mongo.WriteModel{
			&mongo.InsertOneModel{bson.D{{"a", 2}}},
			&mongo.DeleteOneModel{bson.D{{"a", 2}}, nil},
		}
		_, err := mt.Coll.BulkWrite(mtest.Background, models)
		assert.NotNil(mt, err, "expected non-nil error, got nil")

		we, ok := err.(mongo.BulkWriteException)
		assert.True(mt, ok, "expected mongo.BulkWriteException, got %T", err)
		assert.True(mt, we.HasErrorLabel(label), "expected error to have label: %v", label)
	})
}

func TestHintErrors(t *testing.T) {
	mtOpts := mtest.NewOptions().MaxServerVersion("3.2").CreateClient(false)
	mt := mtest.New(t, mtOpts)
	defer mt.Close()

	expected := errors.New("the 'hint' command parameter requires a minimum server wire version of 5")
	mt.Run("UpdateMany", func(mt *mtest.T) {

		_, got := mt.Coll.UpdateMany(mtest.Background, bson.D{{"a", 1}}, bson.D{{"$inc", bson.D{{"a", 1}}}},
			options.Update().SetHint("_id_"))
		assert.NotNil(mt, got, "expected non-nil error, got nil")
		assert.Equal(mt, got, expected, "expected: %v got: %v", expected, got)
	})

	mt.Run("ReplaceOne", func(mt *mtest.T) {

		_, got := mt.Coll.ReplaceOne(mtest.Background, bson.D{{"a", 1}}, bson.D{{"a", 2}},
			options.Replace().SetHint("_id_"))
		assert.NotNil(mt, got, "expected non-nil error, got nil")
		assert.Equal(mt, got, expected, "expected: %v got: %v", expected, got)
	})

	mt.Run("BulkWrite", func(mt *mtest.T) {
		models := []mongo.WriteModel{
			&mongo.InsertOneModel{bson.D{{"_id", 2}}},
			&mongo.ReplaceOneModel{Filter: bson.D{{"_id", 2}}, Replacement: bson.D{{"a", 2}}, Hint: "_id_"},
		}
		_, got := mt.Coll.BulkWrite(mtest.Background, models)
		assert.NotNil(mt, got, "expected non-nil error, got nil")
		assert.Equal(mt, got, expected, "expected: %v got: %v", expected, got)
	})
}
