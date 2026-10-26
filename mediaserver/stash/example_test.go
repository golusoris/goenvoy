package stash_test

import (
	"context"
	"fmt"
	"log"

	"github.com/golusoris/goenvoy/mediaserver/stash/v2"
)

func Example() {
	// Create a new Stash client
	client, err := stash.New("http://localhost:9999/graphql", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get system status
	status, err := client.GetSystemStatus(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Stash status: %s\n", status.Status)

	// Find scenes
	scenes, count, err := client.FindScenes(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total scenes: %d\n", count)
	_ = scenes
}

func ExampleClient_PerformerCreate() {
	client, err := stash.New("http://localhost:9999/graphql", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	performer, err := client.PerformerCreate(ctx, stash.PerformerCreateInput{
		Name: "Jane Doe",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created performer: %s (ID: %s)\n", performer.Name, performer.ID)
}

func ExampleClient_PerformerUpdate() {
	client, err := stash.New("http://localhost:9999/graphql", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	name := "Jane Smith"
	performer, err := client.PerformerUpdate(ctx, stash.PerformerUpdateInput{
		ID:   "1",
		Name: &name,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated performer: %s\n", performer.Name)
}

func ExampleClient_SceneUpdate() {
	client, err := stash.New("http://localhost:9999/graphql", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	title := "Scene Title"
	scene, err := client.SceneUpdate(ctx, stash.SceneUpdateInput{
		ID:           "1",
		Title:        &title,
		PerformerIDs: []string{"p1", "p2"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated scene: %s\n", scene.Title)
}
