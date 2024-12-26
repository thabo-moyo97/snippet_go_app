package database

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestModel struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Value string `db:"value"`
}

func setupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, *DatabaseOperations) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	ops := NewDatabaseOperations(sqlxDB, "test_table", []string{"name", "value"})

	return sqlxDB, mock, ops
}

func TestGetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		mock    func(mock sqlmock.Sqlmock)
		want    *TestModel
		wantErr error
	}{
		{
			name: "successful retrieval",
			id:   1,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "value"}).
					AddRow(1, "test", "value")
				mock.ExpectQuery("SELECT \\* FROM test_table WHERE id = \\?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: &TestModel{ID: 1, Name: "test", Value: "value"},
		},
		{
			name: "record not found",
			id:   2,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM test_table WHERE id = \\?").
					WithArgs(2).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: ErrNoRecord,
		},
		{
			name: "database error",
			id:   3,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM test_table WHERE id = \\?").
					WithArgs(3).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
		{
			name: "zero id",
			id:   0,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM test_table WHERE id = \\?").
					WithArgs(0).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: ErrNoRecord,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, ops := setupTestDB(t)
			defer db.Close()

			tt.mock(mock)

			var got TestModel
			err := ops.GetByID(tt.id, &got)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, &got)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestListPagination(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		offset  int
		mock    func(mock sqlmock.Sqlmock)
		want    []TestModel
		wantErr error
	}{
		{
			name:   "successful list",
			limit:  2,
			offset: 0,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "value"}).
					AddRow(1, "test1", "value1").
					AddRow(2, "test2", "value2")
				mock.ExpectQuery("SELECT \\* FROM test_table LIMIT \\? OFFSET \\?").
					WithArgs(2, 0).
					WillReturnRows(rows)
			},
			want: []TestModel{
				{ID: 1, Name: "test1", Value: "value1"},
				{ID: 2, Name: "test2", Value: "value2"},
			},
		},
		{
			name:   "empty result",
			limit:  2,
			offset: 10,
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "value"})
				mock.ExpectQuery("SELECT \\* FROM test_table LIMIT \\? OFFSET \\?").
					WithArgs(2, 10).
					WillReturnRows(rows)
			},
			want: []TestModel(nil),
		},
		{
			name:   "negative limit",
			limit:  -1,
			offset: 0,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM test_table LIMIT \\? OFFSET \\?").
					WithArgs(-1, 0).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
		{
			name:   "negative offset",
			limit:  10,
			offset: -1,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM test_table LIMIT \\? OFFSET \\?").
					WithArgs(10, -1).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
		{
			name:   "database error",
			limit:  10,
			offset: 0,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM test_table LIMIT \\? OFFSET \\?").
					WithArgs(10, 0).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, ops := setupTestDB(t)
			defer db.Close()

			tt.mock(mock)

			var got []TestModel
			err := ops.ListPagination(tt.limit, tt.offset, &got)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestInsert(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]interface{}
		mock    func(mock sqlmock.Sqlmock)
		want    int
		wantErr error
	}{
		{
			name: "successful insert",
			data: map[string]interface{}{
				"name":  "test",
				"value": "test_value",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO test_table").
					WithArgs("test", "test_value").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want: 1,
		},
		{
			name: "insert error",
			data: map[string]interface{}{
				"name":  "test",
				"value": "test_value",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO test_table").
					WithArgs("test", "test_value").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
		{
			name: "empty data",
			data: map[string]interface{}{},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO test_table").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
		{
			name: "nil value in data",
			data: map[string]interface{}{
				"name":  nil,
				"value": "test_value",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO test_table").
					WithArgs(nil, "test_value").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want: 1,
		},
		{
			name: "LastInsertId error",
			data: map[string]interface{}{
				"name": "test",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO test_table").
					WithArgs("test").
					WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))
			},
			wantErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, ops := setupTestDB(t)
			defer db.Close()

			tt.mock(mock)

			got, err := ops.Insert(tt.data)

			if tt.wantErr != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		data    map[string]interface{}
		mock    func(mock sqlmock.Sqlmock)
		want    bool
		wantErr error
	}{
		{
			name: "successful update",
			id:   1,
			data: map[string]interface{}{
				"name":  "updated",
				"value": "updated_value",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE test_table").
					WithArgs("updated", "updated_value", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			want: true,
		},
		{
			name: "update error",
			id:   1,
			data: map[string]interface{}{
				"name": "test",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE test_table").
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantErr: sql.ErrConnDone,
		},
		{
			name: "begin transaction error",
			id:   1,
			data: map[string]interface{}{
				"name": "test",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
		{
			name: "commit error",
			id:   1,
			data: map[string]interface{}{
				"name": "test",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE test_table").
					WithArgs("test", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit().WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
		{
			name: "zero id",
			id:   0,
			data: map[string]interface{}{
				"name": "test",
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE test_table").
					WithArgs("test", 0).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, ops := setupTestDB(t)
			defer db.Close()

			tt.mock(mock)

			got, err := ops.Update(tt.id, tt.data)

			if tt.wantErr != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		mock    func(mock sqlmock.Sqlmock)
		wantErr error
	}{
		{
			name: "successful delete",
			id:   1,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM test_table WHERE id = \\?").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "record not found",
			id:   2,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM test_table WHERE id = \\?").
					WithArgs(2).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: ErrNoRecord,
		},
		{
			name: "zero id",
			id:   0,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM test_table WHERE id = \\?").
					WithArgs(0).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: ErrNoRecord,
		},
		{
			name: "RowsAffected error",
			id:   1,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM test_table WHERE id = \\?").
					WithArgs(1).
					WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))
			},
			wantErr: sql.ErrConnDone,
		},
		{
			name: "database error",
			id:   1,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM test_table WHERE id = \\?").
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, ops := setupTestDB(t)
			defer db.Close()

			tt.mock(mock)

			err := ops.Delete(tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		mock    func(mock sqlmock.Sqlmock)
		want    bool
		wantErr error
	}{
		{
			name: "record exists",
			id:   1,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
			},
			want: true,
		},
		{
			name: "record does not exist",
			id:   2,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(0))
			},
			want: false,
		},
		{
			name: "database error",
			id:   1,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			wantErr: sql.ErrConnDone,
		},
		{
			name: "zero id",
			id:   0,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT EXISTS").
					WithArgs(0).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(0))
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, ops := setupTestDB(t)
			defer db.Close()

			tt.mock(mock)

			got, err := ops.Exists(tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetFillableFields(t *testing.T) {
	db, _, ops := setupTestDB(t)
	defer db.Close()

	want := []string{"name", "value"}
	got := ops.GetFillableFields()

	assert.Equal(t, want, got)
}

type anyDriver struct{}

func (d *anyDriver) Open(name string) (driver.Conn, error) {
	return nil, fmt.Errorf("unsupported driver")
}

// TestDatabaseOperations tests the DatabaseOperations struct
func TestDatabaseOperations(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(mock sqlmock.Sqlmock)
		run           func(ops *DatabaseOperations, t *testing.T)
		expectedError error
	}{
		{
			name: "GetByID - Record Found",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email"}).
					AddRow(1, "John Doe", "john.doe@example.com")
				mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			run: func(ops *DatabaseOperations, t *testing.T) {
				var user struct {
					ID    int
					Name  string
					Email string
				}
				err := ops.GetByID(1, &user)
				assert.NoError(t, err)
				assert.Equal(t, user.ID, 1)
				assert.Equal(t, user.Name, "John Doe")
				assert.Equal(t, user.Email, "john.doe@example.com")
			},
			expectedError: nil,
		},
		{
			name: "GetByID - No Record Found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\?").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			run: func(ops *DatabaseOperations, t *testing.T) {
				var user struct {
					ID    int
					Name  string
					Email string
				}
				err := ops.GetByID(1, &user)
				assert.ErrorIs(t, err, ErrNoRecord)
			},
			expectedError: ErrNoRecord,
		},
		{
			name: "ListPagination - Success",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email"}).
					AddRow(1, "John Doe", "john.doe@example.com").
					AddRow(2, "Jane Smith", "jane.smith@example.com")
				mock.ExpectQuery("SELECT \\* FROM users LIMIT \\? OFFSET \\?").
					WithArgs(2, 0).
					WillReturnRows(rows)
			},
			run: func(ops *DatabaseOperations, t *testing.T) {
				var users []struct {
					ID    int
					Name  string
					Email string
				}
				err := ops.ListPagination(2, 0, &users)
				assert.NoError(t, err)
				assert.Len(t, users, 2)
				assert.Equal(t, users[0].ID, 1)
				assert.Equal(t, users[0].Name, "John Doe")
				assert.Equal(t, users[0].Email, "john.doe@example.com")
				assert.Equal(t, users[1].ID, 2)
				assert.Equal(t, users[1].Name, "Jane Smith")
				assert.Equal(t, users[1].Email, "jane.smith@example.com")
			},
			expectedError: nil,
		},
		{
			name: "Exists - Record Found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE id = \\?\\)").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			run: func(ops *DatabaseOperations, t *testing.T) {
				exists, err := ops.Exists(1)
				assert.NoError(t, err)
				assert.True(t, exists)
			},
			expectedError: nil,
		},
		{
			name: "Exists - No Record Found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE id = \\?\\)").
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			},
			run: func(ops *DatabaseOperations, t *testing.T) {
				exists, err := ops.Exists(2)
				assert.NoError(t, err)
				assert.False(t, exists)
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			ops := NewDatabaseOperations(sqlxDB, "users", []string{"name", "email"})

			if tc.setupMock != nil {
				tc.setupMock(mock)
			}

			tc.run(ops, t)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
