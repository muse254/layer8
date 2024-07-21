package repository

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"globe-and-citizen/layer8/server/resource_server/dto"
	"globe-and-citizen/layer8/server/resource_server/interfaces"
	"globe-and-citizen/layer8/server/resource_server/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"regexp"
	"testing"
	"time"
)

const id uint = 1

const userId uint = 1
const username = "test_username"
const userFirstName = "first_name"
const userLastName = "last_name"
const userSalt = "user_salt"
const userPassword = "user_password"
const userDisplayName = "display_name"

const verificationCode = "123456"

const clientId = "1"
const clientUsername = "client_username"
const clientName = "test_client"
const clientSecret = "client_secret"
const redirectUri = "https://gcitizen.com/callback"
const clientSalt = "client_salt"
const clientPassword = "client_password"

var timestamp = time.Date(2024, time.May, 24, 14, 0, 0, 0, time.UTC)

var emailProof = []byte("AbcdfTs")

var mockDB *sql.DB
var mock sqlmock.Sqlmock
var err error
var db *gorm.DB
var repository interfaces.IRepository

func SetUp(t *testing.T) {
	mockDB, mock, err = sqlmock.New()
	if err != nil {
		t.Fatal("Failed to create mock DB:", err)
	}

	db, err = gorm.Open(
		postgres.New(
			postgres.Config{
				Conn: mockDB,
			},
		),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatal("Failed to connect to mock DB:", err)
	}

	repository = NewRepository(db)
}

func TestRegisterUser(t *testing.T) {
	// Create a new mock DB and a GORM database connection
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatal("Failed to create mock DB:", err)
	}
	defer mockDB.Close()

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect to mock DB:", err)
	}

	// Create the user repository with the mock database connection
	repo := NewRepository(db)

	// Define a test user DTO
	testUser := dto.RegisterUserDTO{
		Username:    "test_user",
		FirstName:   "Test",
		LastName:    "User",
		Password:    "TestPass123",
		Country:     "Unknown",
		DisplayName: "user",
	}

	// Call the RegisterUser function
	repo.RegisterUser(testUser)
}

func TestRegisterClient(t *testing.T) {
	// Create a new mock DB and a GORM database connection
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatal("Failed to create mock DB:", err)
	}
	defer mockDB.Close()

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect to mock DB:", err)
	}

	// Create the client repository with the mock database connection
	repo := NewRepository(db)

	// Define a test client DTO
	testClient := dto.RegisterClientDTO{
		Name:        "testclient",
		RedirectURI: "https://gcitizen.com/callback",
	}

	// Call the RegisterClient function
	repo.RegisterClient(testClient)
}

func TestGetClientData_ClientDoesNotExist(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "clients" WHERE name = $1 ORDER BY "clients"."id" LIMIT 1`),
	).WithArgs(
		clientName,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "secret", "name", "redirect_uri"},
		),
	)

	_, err := repository.GetClientData(clientName)

	assert.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetClientData_Success(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "clients" WHERE name = $1 ORDER BY "clients"."id" LIMIT 1`),
	).WithArgs(
		clientName,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "secret", "name", "redirect_uri", "username", "salt", "password"},
		).AddRow(
			clientId, clientSecret, clientName, redirectUri, clientUsername, clientSalt, clientPassword,
		),
	)

	client, err := repository.GetClientData(clientName)

	assert.Nil(t, err)

	assert.Equal(t, clientId, client.ID)
	assert.Equal(t, clientSecret, client.Secret)
	assert.Equal(t, clientName, client.Name)
	assert.Equal(t, redirectUri, client.RedirectURI)
	assert.Equal(t, clientUsername, client.Username)
	assert.Equal(t, clientSalt, client.Salt)
	assert.Equal(t, clientPassword, client.Password)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestLoginPreCheckUser_UserDoesNotExist(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."id" LIMIT 1`),
	).WithArgs(
		username,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "username", "first_name", "last_name", "password", "salt"},
		),
	)

	_, _, err := repository.LoginPreCheckUser(
		dto.LoginPrecheckDTO{
			Username: username,
		},
	)

	assert.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestLoginPreCheckUser_Success(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."id" LIMIT 1`),
	).WithArgs(
		username,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "username", "first_name", "last_name", "password", "salt"},
		).AddRow(
			userId, username, userFirstName, userLastName, userPassword, userSalt,
		),
	)

	precheckedUsername, precheckedSalt, err := repository.LoginPreCheckUser(
		dto.LoginPrecheckDTO{
			Username: username,
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, username, precheckedUsername)
	assert.Equal(t, userSalt, precheckedSalt)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestProfileUser_UserIsNotFoundInTheUsersTable(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT 1`),
	).WithArgs(
		userId,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "username", "first_name", "last_name", "password", "salt"},
		),
	)

	_, _, err := repository.ProfileUser(userId)

	assert.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestProfileUser_Success(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT 1`),
	).WithArgs(
		userId,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "username", "first_name", "last_name", "password", "salt"},
		).AddRow(
			userId, username, userFirstName, userLastName, userPassword, userSalt,
		),
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "user_metadata" WHERE user_id = $1`),
	).WithArgs(
		userId,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "user_id", "key", "value"},
		).AddRow(
			1, userId, "email_verified", "true",
		).AddRow(
			2, userId, "display_name", "user",
		).AddRow(
			3, userId, "country", "Unknown",
		),
	)

	user, userMetadata, err := repository.ProfileUser(userId)

	assert.Nil(t, err)
	assert.Equal(t, userId, user.ID)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, userFirstName, user.FirstName)
	assert.Equal(t, userLastName, user.LastName)
	assert.Equal(t, userPassword, user.Password)
	assert.Equal(t, userSalt, user.Salt)

	for _, metadata := range userMetadata {
		assert.Equal(t, userId, metadata.UserID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestUpdateDisplayName_TableUpdateFails(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectBegin()

	mock.ExpectExec(
		regexp.QuoteMeta(`UPDATE "user_metadata" SET "value"=$1,"updated_at"=$2 WHERE user_id = $3 AND key = $4`),
	).WithArgs(
		userDisplayName, sqlmock.AnyArg(), userId, "display_name",
	).WillReturnError(
		fmt.Errorf("failed to update display name"),
	)

	mock.ExpectRollback()

	err := repository.UpdateDisplayName(
		userId,
		dto.UpdateDisplayNameDTO{
			DisplayName: userDisplayName,
		},
	)

	assert.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestUpdateDisplayName_Success(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectBegin()

	mock.ExpectExec(
		regexp.QuoteMeta(`UPDATE "user_metadata" SET "value"=$1,"updated_at"=$2 WHERE user_id = $3 AND key = $4`),
	).WithArgs(
		userDisplayName, sqlmock.AnyArg(), userId, "display_name",
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)

	mock.ExpectCommit()

	err := repository.UpdateDisplayName(
		userId,
		dto.UpdateDisplayNameDTO{
			DisplayName: userDisplayName,
		},
	)

	assert.Nil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestLoginClient_ClientNotFound(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "clients" WHERE username = $1 ORDER BY "clients"."id" LIMIT 1`),
	).WithArgs(
		clientUsername,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "secret", "username", "password"},
		),
	)

	_, err := repository.LoginClient(
		dto.LoginClientDTO{
			Username: clientUsername,
			Password: clientPassword,
		},
	)

	assert.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestLoginClient_Success(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "clients" WHERE username = $1 ORDER BY "clients"."id" LIMIT 1`),
	).WithArgs(
		clientUsername,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "secret", "username", "password"},
		).AddRow(
			clientId, clientSecret, clientUsername, clientPassword,
		),
	)

	client, err := repository.LoginClient(
		dto.LoginClientDTO{
			Username: clientUsername,
			Password: clientPassword,
		},
	)

	assert.Nil(t, err)

	assert.Equal(t, clientId, client.ID)
	assert.Equal(t, clientUsername, client.Username)
	assert.Equal(t, clientSecret, client.Secret)
	assert.Equal(t, clientPassword, client.Password)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestProfileClient_ClientNotFound(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "clients" WHERE username = $1 ORDER BY "clients"."id" LIMIT 1`),
	).WithArgs(
		clientUsername,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "secret", "name", "redirect_uri", "username"},
		),
	)

	_, err := repository.ProfileClient(clientUsername)

	assert.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestProfileClient_Success(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "clients" WHERE username = $1 ORDER BY "clients"."id" LIMIT 1`),
	).WithArgs(
		clientUsername,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "secret", "name", "redirect_uri", "username"},
		).AddRow(
			clientId, clientSecret, clientName, redirectUri, clientUsername,
		),
	)

	client, err := repository.ProfileClient(clientUsername)

	assert.Nil(t, err)

	assert.Equal(t, clientId, client.ID)
	assert.Equal(t, clientUsername, client.Username)
	assert.Equal(t, clientName, client.Name)
	assert.Equal(t, clientSecret, client.Secret)
	assert.Equal(t, redirectUri, client.RedirectURI)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestFindUser_UserDoesNotExist(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT 1`),
	).WithArgs(
		userId,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "username", "password", "first_name",
				"last_name", "salt", "email_proof", "verification_code"},
		),
	)

	_, err := repository.FindUser(userId)

	assert.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestFindUser_Success(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "users" WHERE id = $1 ORDER BY "users"."id" LIMIT 1`),
	).WithArgs(
		userId,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "username", "password", "first_name",
				"last_name", "salt", "email_proof", "verification_code"},
		).AddRow(userId, username, "password", "first_name",
			"last_name", "salt", "", "",
		),
	)

	user, err := repository.FindUser(userId)

	if err != nil {
		t.Fatalf("Error while retrieving user: %v", err)
	}

	assert.Equal(t, userId, user.ID)
	assert.Equal(t, username, user.Username)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetEmailVerificationData_VerificationDataForTheGivenUserDoesNotExist(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "email_verification_data" WHERE user_id = $1 ORDER BY "email_verification_data"."id" LIMIT 1`,
		),
	).WithArgs(
		userId,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "user_id", "verification_code", "expires_at"},
		),
	)

	_, err := repository.GetEmailVerificationData(userId)

	assert.NotNil(t, err)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetEmailVerificationData_Success(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "email_verification_data" WHERE user_id = $1 ORDER BY "email_verification_data"."id" LIMIT 1`,
		),
	).WithArgs(
		userId,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "user_id", "verification_code", "expires_at"},
		).AddRow(
			id, userId, verificationCode, timestamp,
		),
	)

	data, err := repository.GetEmailVerificationData(userId)

	if err != nil {
		t.Fatalf("Error while getting email verification data: %v", err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}

	assert.Equal(t, userId, data.UserId)
	assert.Equal(t, verificationCode, data.VerificationCode)
	assert.Equal(t, timestamp, data.ExpiresAt)
}

func TestSaveEmailVerificationData_RowWithUserIdDoesNotExist(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectBegin()

	mock.ExpectQuery(
		regexp.QuoteMeta(`SELECT * FROM "email_verification_data" WHERE "email_verification_data"."user_id" = $1 ORDER BY "email_verification_data"."id" LIMIT 1`),
	).WithArgs(
		userId,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "user_id", "verification_code", "expires_at"},
		),
	)

	mock.ExpectQuery(
		regexp.QuoteMeta(
			`INSERT INTO "email_verification_data" ("user_id","verification_code","expires_at") VALUES ($1,$2,$3) RETURNING "id"`,
		),
	).WithArgs(
		userId, verificationCode, timestamp,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "user_id", "verification_code", "expires_at"},
		).AddRow(
			1, userId, verificationCode, timestamp,
		),
	)

	mock.ExpectCommit()

	err = repository.SaveEmailVerificationData(
		models.EmailVerificationData{
			UserId:           userId,
			VerificationCode: verificationCode,
			ExpiresAt:        timestamp,
		},
	)

	if err != nil {
		t.Fatalf("Error while saving email verification data: %v", err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestSaveEmailVerificationData_RowWithUserIdExists(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectBegin()

	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "email_verification_data" WHERE "email_verification_data"."user_id" = $1 ORDER BY "email_verification_data"."id" LIMIT 1`,
		),
	).WithArgs(
		userId,
	).WillReturnRows(
		sqlmock.NewRows(
			[]string{"id", "user_id", "verification_code", "expires_at"},
		).AddRow(
			id, userId, verificationCode, timestamp,
		),
	)

	mock.ExpectExec(
		regexp.QuoteMeta(
			`UPDATE "email_verification_data" SET "expires_at"=$1,"user_id"=$2,"verification_code"=$3 WHERE "email_verification_data"."user_id" = $4 AND "id" = $5`,
		),
	).WithArgs(
		timestamp, userId, verificationCode, userId, id,
	).WillReturnResult(sqlmock.NewResult(2, 1))

	mock.ExpectCommit()

	err = repository.SaveEmailVerificationData(
		models.EmailVerificationData{
			UserId:           userId,
			VerificationCode: verificationCode,
			ExpiresAt:        timestamp,
		},
	)

	if err != nil {
		t.Fatalf("Error while saving email verification data: %v", err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestSaveProofOfEmailVerification_UsersTableUpdateFails(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectBegin()

	mock.ExpectExec(
		regexp.QuoteMeta(`UPDATE "users" SET "email_proof"=$1,"verification_code"=$2 WHERE id = $3`),
	).WithArgs(
		emailProof, verificationCode, userId,
	).WillReturnError(
		fmt.Errorf(""),
	)

	mock.ExpectRollback()

	err = repository.SaveProofOfEmailVerification(userId, verificationCode, emailProof)

	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestSaveProofOfEmailVerification_DeletingEmailVerificationDataFails(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectBegin()

	mock.ExpectExec(
		regexp.QuoteMeta(`UPDATE "users" SET "email_proof"=$1,"verification_code"=$2 WHERE id = $3`),
	).WithArgs(
		emailProof, verificationCode, userId,
	).WillReturnResult(
		sqlmock.NewResult(0, 1),
	)

	mock.ExpectExec(
		regexp.QuoteMeta(`DELETE FROM "email_verification_data" WHERE user_id = $1`),
	).WithArgs(
		userId,
	).WillReturnError(
		fmt.Errorf(""),
	)

	mock.ExpectRollback()

	err = repository.SaveProofOfEmailVerification(userId, verificationCode, emailProof)

	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestSaveProofOfEmailVerification_SettingUserStatusAsVerifiedFails(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectBegin()

	mock.ExpectExec(
		regexp.QuoteMeta(`UPDATE "users" SET "email_proof"=$1,"verification_code"=$2 WHERE id = $3`),
	).WithArgs(
		emailProof, verificationCode, userId,
	).WillReturnResult(
		sqlmock.NewResult(0, 1),
	)

	mock.ExpectExec(
		regexp.QuoteMeta(`DELETE FROM "email_verification_data" WHERE user_id = $1`),
	).WithArgs(
		userId,
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)

	mock.ExpectExec(
		regexp.QuoteMeta(`UPDATE "user_metadata" SET "value"=$1,"updated_at"=$2 WHERE user_id = $3 AND key = $4`),
	).WithArgs(
		"true",
		sqlmock.AnyArg(),
		userId,
		"email_verified",
	).WillReturnError(
		fmt.Errorf(""),
	)

	mock.ExpectRollback()

	err = repository.SaveProofOfEmailVerification(userId, verificationCode, emailProof)

	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestSaveProofOfEmailVerification_Success(t *testing.T) {
	SetUp(t)
	defer mockDB.Close()

	mock.ExpectBegin()

	mock.ExpectExec(
		regexp.QuoteMeta(`UPDATE "users" SET "email_proof"=$1,"verification_code"=$2 WHERE id = $3`),
	).WithArgs(
		emailProof, verificationCode, userId,
	).WillReturnResult(
		sqlmock.NewResult(0, 1),
	)

	mock.ExpectExec(
		regexp.QuoteMeta(`DELETE FROM "email_verification_data" WHERE user_id = $1`),
	).WithArgs(
		userId,
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)

	mock.ExpectExec(
		regexp.QuoteMeta(`UPDATE "user_metadata" SET "value"=$1,"updated_at"=$2 WHERE user_id = $3 AND key = $4`),
	).WithArgs(
		"true",
		sqlmock.AnyArg(),
		userId,
		"email_verified",
	).WillReturnResult(
		sqlmock.NewResult(0, 1),
	)

	mock.ExpectCommit()

	err = repository.SaveProofOfEmailVerification(userId, verificationCode, emailProof)

	assert.Nil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
