package tests

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vladovsiychuk/auth-service-go/configs"
	"github.com/vladovsiychuk/auth-service-go/internal/auth"
	pgDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostgresRepository(t *testing.T) {
	ctx := context.Background()

	dbName := "testdb"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:latest",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	host, _ := postgresContainer.Host(ctx)
	port, _ := postgresContainer.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, dbUser, dbPassword, dbName, port.Port())
	postgresDB, err := gorm.Open(pgDriver.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	configs.SetupDbMigration(postgresDB)

	/*
	*
	* Test Keys Repository
	*
	 */

	keysRepository := auth.NewKeyRepository(postgresDB)

	newKeys := auth.CreateKeys()

	if err := keysRepository.Update(newKeys); err != nil {
		panic(err)
	}

	savedKeys, err := keysRepository.GetKeys()
	if err != nil {
		panic(err)
	}

	privateKey, err := savedKeys.GetPrivateKey()
	if err != nil {
		panic(err)
	}

	assert.NotNil(t, privateKey)

	/*
	*
	* Session Tokens Repository
	*
	 */

	sessionTokensRepository := auth.NewSessionTokenRepository(postgresDB)

	newSessionTokenI := auth.CreateSessionToken("test@mail.com")

	if err := sessionTokensRepository.Create(newSessionTokenI); err != nil {
		panic(err)
	}

	newSessionToken := newSessionTokenI.(*auth.SessionToken)

	savedSessionTokenI, err := sessionTokensRepository.FindById(newSessionToken.Id)
	if err != nil {
		panic(err)
	}

	savedSessionToken := savedSessionTokenI.(*auth.SessionToken)

	assert.NotNil(t, savedSessionToken.Id)
	assert.False(t, savedSessionToken.ExpiresAt.Before(time.Now()))

	if err := sessionTokensRepository.Delete(savedSessionTokenI); err != nil {
		panic(err)
	}

	_, err = sessionTokensRepository.FindById(newSessionToken.Id)

	assert.NotNil(t, err)
}
