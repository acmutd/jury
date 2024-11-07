package database

import (
	"context"
	"server/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetOptions gets the options from the database
func GetOptions(db *mongo.Database, ctx context.Context) (*models.Options, error) {
	var options models.Options
	err := db.Collection("options").FindOne(ctx, gin.H{}).Decode(&options)

	// If options does not exist, create it
	if err == mongo.ErrNoDocuments {
		options = *models.NewOptions()
		_, err = db.Collection("options").InsertOne(context.Background(), options)
		return &options, err
	}

	return &options, err
}

// UpdateOptions updates the options in the database
func UpdateOptions(db *mongo.Database, ctx context.Context, options *models.OptionalOptions) error {
	update := gin.H{}

	if options.JudgingTimer != nil {
		update["judging_timer"] = *options.JudgingTimer
	}
	if options.MinViews != nil {
		update["min_views"] = *options.MinViews
	}
	if options.ClockSync != nil {
		update["clock_sync"] = *options.ClockSync
	}
	if options.JudgeTracks != nil {
		update["judge_tracks"] = *options.JudgeTracks
	}
	if options.Tracks != nil {
		update["tracks"] = *options.Tracks
	}
	if options.MultiGroup != nil {
		update["multi_group"] = *options.MultiGroup
	}
	if options.GroupSizes != nil {
		update["group_sizes"] = *options.GroupSizes
	}
	if options.SwitchingMode != nil {
		update["switching_mode"] = *options.SwitchingMode
	}
	if options.AutoSwitchProp != nil {
		update["auto_switch_prop"] = *options.AutoSwitchProp
	}

	_, err := db.Collection("options").UpdateOne(ctx, gin.H{}, gin.H{"$set": update})
	return err
}

// UpdateClockConditional updates the clock in the database if clock sync is enabled
func UpdateClockConditional(db *mongo.Database, ctx context.Context, clock *models.ClockState) error {
	// Get options
	options, err := GetOptions(db, ctx)
	if err != nil {
		return err
	}

	// If clock sync is not enabled, don't sync the clock
	if !options.ClockSync {
		return nil
	}

	_, err = db.Collection("options").UpdateOne(ctx, gin.H{}, gin.H{"$set": gin.H{"clock": clock}})
	return err
}

// UpdateClock updates the clock in the database
func UpdateClock(db *mongo.Database, clock *models.ClockState) error {
	_, err := db.Collection("options").UpdateOne(context.Background(), gin.H{}, gin.H{"$set": gin.H{"clock": clock}})
	return err
}

// UpdateNumGroups will update the number of groups and resize the group sizes if necessary
func UpdateNumGroups(db *mongo.Database, ctx context.Context, numGroups int64) error {
	// Get options
	options, err := GetOptions(db, ctx)
	if err != nil {
		return err
	}

	// Resize group sizes if necessary
	if numGroups < options.NumGroups {
		options.GroupSizes = options.GroupSizes[:numGroups-1]
	} else if numGroups > options.NumGroups {
		for i := options.NumGroups; i < numGroups-1; i++ {
			options.GroupSizes = append(options.GroupSizes, 30)
		}
	}

	// Reassign group numbers to all projects
	ReassignAllGroupNums(db, ctx, options)

	_, err = db.Collection("options").UpdateOne(ctx, gin.H{}, gin.H{"$set": gin.H{"num_groups": numGroups, "group_sizes": options.GroupSizes}})
	return err
}

// IncrementManualSwitches increments the manual switches in the database
func IncrementManualSwitches(db *mongo.Database, ctx context.Context) error {
	_, err := db.Collection("options").UpdateOne(ctx, gin.H{}, gin.H{"$inc": gin.H{"manual_switches": 1}})
	return err
}
