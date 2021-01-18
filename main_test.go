package main

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, assume env variables are set.")
	}

	guacamoleEndpoint := os.Getenv("GUACAMOLE_ENDPOINT")
	guacamoleUsername := os.Getenv("GUACAMOLE_USERNAME")
	guacamolePassword := os.Getenv("GUACAMOLE_PASSWORD")
	//guacamoleDataSource := os.Getenv("GUACAMOLE_DATASOURCE")

	token, err := getToken(guacamoleEndpoint, guacamoleUsername, guacamolePassword)
	assert.Nil(t, err)

	releaseToken(guacamoleEndpoint, token)
}

func TestActiveConnections(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, assume env variables are set.")
	}

	guacamoleEndpoint := os.Getenv("GUACAMOLE_ENDPOINT")
	guacamoleUsername := os.Getenv("GUACAMOLE_USERNAME")
	guacamolePassword := os.Getenv("GUACAMOLE_PASSWORD")
	guacamoleDataSource := os.Getenv("GUACAMOLE_DATASOURCE")

	token, err := getToken(guacamoleEndpoint, guacamoleUsername, guacamolePassword)
	assert.Nil(t, err)

	activeConnections, err := getActiveConnections(guacamoleEndpoint, token, guacamoleDataSource)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, activeConnections, 0)

	releaseToken(guacamoleEndpoint, token)
}

func TestConnectionHistory(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, assume env variables are set.")
	}

	guacamoleEndpoint := os.Getenv("GUACAMOLE_ENDPOINT")
	guacamoleUsername := os.Getenv("GUACAMOLE_USERNAME")
	guacamolePassword := os.Getenv("GUACAMOLE_PASSWORD")
	guacamoleDataSource := os.Getenv("GUACAMOLE_DATASOURCE")

	token, err := getToken(guacamoleEndpoint, guacamoleUsername, guacamolePassword)
	assert.Nil(t, err)

	connectionHistory, err := getConnectionHistory(guacamoleEndpoint, token, guacamoleDataSource)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, connectionHistory, 0)

	releaseToken(guacamoleEndpoint, token)
}

func TestUsers(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, assume env variables are set.")
	}

	guacamoleEndpoint := os.Getenv("GUACAMOLE_ENDPOINT")
	guacamoleUsername := os.Getenv("GUACAMOLE_USERNAME")
	guacamolePassword := os.Getenv("GUACAMOLE_PASSWORD")
	guacamoleDataSource := os.Getenv("GUACAMOLE_DATASOURCE")

	token, err := getToken(guacamoleEndpoint, guacamoleUsername, guacamolePassword)
	assert.Nil(t, err)

	users, err := getUsers(guacamoleEndpoint, token, guacamoleDataSource)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, users, 0)

	releaseToken(guacamoleEndpoint, token)
}
