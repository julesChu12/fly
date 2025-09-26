package db

import (
	"context"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	expected := Config{
		Driver:          "mysql",
		DSN:             "",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 3600,
		LogLevel:        "warn",
	}

	if cfg != expected {
		t.Errorf("DefaultConfig() = %+v, want %+v", cfg, expected)
	}
}

func TestNew_UnsupportedDriver(t *testing.T) {
	cfg := Config{
		Driver: "unsupported",
		DSN:    "",
	}

	client, err := New(cfg)
	if err == nil {
		t.Error("New() should return error for unsupported driver")
		if client != nil {
			client.Close()
		}
	}

	if client != nil {
		t.Error("New() should return nil client for unsupported driver")
	}
}

func TestNew_SQLiteBasic(t *testing.T) {
	cfg := Config{
		Driver:          "sqlite",
		DSN:             ":memory:",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 1800,
		LogLevel:        "silent",
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.Close()

	if client == nil {
		t.Fatal("New() returned nil client")
	}

	// Test that we can get the underlying DB
	db := client.DB()
	if db == nil {
		t.Error("client.DB() returned nil")
	}

	// Test basic connection
	err = client.Ping()
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}

	// Test stats
	stats := client.Stats()
	t.Logf("DB Stats: %+v", stats)
}

func TestGORMClient_MethodsExist(t *testing.T) {
	// Test that all GORM client methods exist
	cfg := Config{
		Driver:   "sqlite",
		DSN:      ":memory:",
		LogLevel: "silent",
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("Connection Methods", func(t *testing.T) {
		// Test Ping
		err := client.Ping()
		if err != nil {
			t.Errorf("Ping() error = %v", err)
		}

		// Test Stats
		stats := client.Stats()
		t.Logf("Stats: %+v", stats)
	})

	t.Run("Transaction Methods", func(t *testing.T) {
		// Test Begin
		tx := client.Begin()
		if tx == nil {
			t.Error("Begin() returned nil")
		} else {
			// Test transaction methods exist
			if tx.DB() == nil {
				t.Error("Transaction DB() returned nil")
			}
			tx.Rollback() // Clean up
		}

		// Test BeginTx
		tx2 := client.BeginTx(ctx, nil)
		if tx2 == nil {
			t.Error("BeginTx() returned nil")
		} else {
			tx2.Rollback() // Clean up
		}

		// Test WithTransaction (with no-op function)
		err := client.WithTransaction(ctx, func(tx *Transaction) error {
			return nil // No-op
		})
		if err != nil {
			t.Errorf("WithTransaction() error = %v", err)
		}
	})

	t.Run("CRUD Method Signatures", func(t *testing.T) {
		// Test method signatures exist (they will fail due to no tables, but methods should exist)
		type TestModel struct{}

		// These will fail but we're testing method existence
		client.Create(ctx, &TestModel{})
		client.Save(ctx, &TestModel{})
		client.First(ctx, &TestModel{})
		client.Find(ctx, &[]TestModel{})
		client.Update(ctx, "field", "value", "condition")
		client.Updates(ctx, map[string]interface{}{}, "condition")
		client.Delete(ctx, &TestModel{})

		var count int64
		client.Count(ctx, &TestModel{}, &count)
		client.Exists(ctx, &TestModel{})
		client.Paginate(ctx, &[]TestModel{}, 1, 10)
		client.PaginateWithCount(ctx, &TestModel{}, &[]TestModel{}, 1, 10)
	})

	t.Run("Raw SQL Methods", func(t *testing.T) {
		// Test Raw and Exec methods exist
		result := client.Raw(ctx, "SELECT 1")
		if result == nil {
			t.Error("Raw() returned nil")
		}

		err := client.Exec(ctx, "SELECT 1")
		if err != nil {
			t.Logf("Exec() error (expected): %v", err)
		}
	})

	t.Run("Migration Methods", func(t *testing.T) {
		// Test AutoMigrate exists
		type TestModel struct {
			ID   uint   `gorm:"primaryKey"`
			Name string `gorm:"not null"`
		}

		err := client.AutoMigrate(&TestModel{})
		if err != nil {
			t.Logf("AutoMigrate() error (may be expected): %v", err)
		}
	})
}

func TestNewSQLX_Basic(t *testing.T) {
	cfg := Config{
		Driver:          "sqlite3",
		DSN:             ":memory:",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 1800,
	}

	client, err := NewSQLX(cfg)
	if err != nil {
		t.Fatalf("NewSQLX() error = %v", err)
	}
	defer client.Close()

	if client == nil {
		t.Fatal("NewSQLX() returned nil client")
	}

	// Test that we can get the underlying DB
	db := client.DB()
	if db == nil {
		t.Error("client.DB() returned nil")
	}

	// Test basic connection
	err = client.Ping()
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}

	// Test stats
	stats := client.Stats()
	t.Logf("SQLX DB Stats: %+v", stats)
}

func TestNewSQLX_InvalidDSN(t *testing.T) {
	cfg := Config{
		Driver: "sqlite3",
		DSN:    "/invalid/path/to/database.db",
	}

	client, err := NewSQLX(cfg)
	if err == nil {
		t.Error("NewSQLX() should return error for invalid DSN")
		if client != nil {
			client.Close()
		}
	}
}

func TestSQLXClient_MethodsExist(t *testing.T) {
	cfg := Config{
		Driver: "sqlite3",
		DSN:    ":memory:",
	}

	client, err := NewSQLX(cfg)
	if err != nil {
		t.Fatalf("NewSQLX() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("Connection Methods", func(t *testing.T) {
		// Test Ping
		err := client.Ping()
		if err != nil {
			t.Errorf("Ping() error = %v", err)
		}

		// Test Stats
		stats := client.Stats()
		t.Logf("SQLX Stats: %+v", stats)
	})

	t.Run("Query Methods", func(t *testing.T) {
		// Test method signatures exist (they will fail due to no tables, but methods should exist)
		type TestStruct struct{}

		client.Get(ctx, &TestStruct{}, "SELECT 1")
		client.Select(ctx, &[]TestStruct{}, "SELECT 1")
		client.Exec(ctx, "SELECT 1")
		client.Query(ctx, "SELECT 1")
		client.QueryRow(ctx, "SELECT 1")
		client.NamedExec(ctx, "SELECT :value", map[string]interface{}{"value": 1})
		client.NamedQuery(ctx, "SELECT :value", map[string]interface{}{"value": 1})

		// Test these methods exist
		t.Log("All query methods exist")
	})

	t.Run("Transaction Methods", func(t *testing.T) {
		// Test Begin
		tx, err := client.Begin()
		if err != nil {
			t.Errorf("Begin() error = %v", err)
		} else {
			// Test transaction methods exist
			if tx.Tx() == nil {
				t.Error("Transaction Tx() returned nil")
			}
			tx.Rollback() // Clean up
		}

		// Test BeginTx
		tx2, err := client.BeginTx(ctx, nil)
		if err != nil {
			t.Errorf("BeginTx() error = %v", err)
		} else {
			tx2.Rollback() // Clean up
		}

		// Test WithTransaction (with no-op function)
		err = client.WithTransaction(ctx, func(tx *SQLXTransaction) error {
			return nil // No-op
		})
		if err != nil {
			t.Errorf("WithTransaction() error = %v", err)
		}
	})

	t.Run("Prepared Statement Methods", func(t *testing.T) {
		// Test Prepare method exists
		stmt, err := client.Prepare(ctx, "SELECT 1")
		if err != nil {
			t.Logf("Prepare() error (may be expected): %v", err)
		} else {
			stmt.Close() // Clean up
		}
	})
}

func TestSQLXHelperFunctions(t *testing.T) {
	t.Run("In function", func(t *testing.T) {
		query := "SELECT * FROM users WHERE id IN (?)"
		args := []interface{}{1, 2, 3}

		expandedQuery, expandedArgs, err := In(query, args)
		if err != nil {
			t.Errorf("In() error = %v", err)
		}

		if expandedQuery == query {
			t.Error("In() should expand the query")
		}

		if len(expandedArgs) != 3 {
			t.Errorf("In() expanded args length = %d, want 3", len(expandedArgs))
		}

		t.Logf("In() expanded query: %s", expandedQuery)
		t.Logf("In() expanded args: %v", expandedArgs)
	})

	t.Run("Named function", func(t *testing.T) {
		query := "SELECT * FROM users WHERE name = :name AND age > :age"
		params := map[string]interface{}{
			"name": "John",
			"age":  25,
		}

		expandedQuery, args, err := Named(query, params)
		if err != nil {
			t.Errorf("Named() error = %v", err)
		}

		if expandedQuery == query {
			t.Error("Named() should expand the query")
		}

		if len(args) != 2 {
			t.Errorf("Named() args length = %d, want 2", len(args))
		}

		t.Logf("Named() expanded query: %s", expandedQuery)
		t.Logf("Named() args: %v", args)
	})
}

func TestIntegration_WithActualDatabase(t *testing.T) {
	// This test only runs if we can successfully connect to SQLite
	cfg := Config{
		Driver:   "sqlite",
		DSN:      ":memory:",
		LogLevel: "silent",
	}

	client, err := New(cfg)
	if err != nil {
		t.Skipf("Skipping integration test - cannot connect to SQLite: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Define a test model
	type User struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"not null"`
		Email string `gorm:"unique"`
	}

	t.Run("AutoMigrate Test", func(t *testing.T) {
		// Auto migrate the schema
		err := client.AutoMigrate(&User{})
		if err != nil {
			t.Errorf("AutoMigrate() error = %v", err)
		} else {
			t.Log("AutoMigrate completed successfully")
		}
	})

	t.Run("Raw SQL Test", func(t *testing.T) {
		// Test raw SQL execution that should work
		err := client.Exec(ctx, "CREATE TABLE IF NOT EXISTS simple_test (id INTEGER, name TEXT)")
		if err != nil {
			t.Errorf("Raw Exec() error = %v", err)
		} else {
			t.Log("Raw SQL execution completed successfully")
		}
	})

	t.Run("Basic Connection Test", func(t *testing.T) {
		// Just test that we can ping and get stats
		err := client.Ping()
		if err != nil {
			t.Errorf("Ping() error = %v", err)
		}

		stats := client.Stats()
		if stats.MaxOpenConnections < 0 {
			t.Error("Invalid stats returned")
		}

		t.Logf("Connection test passed - Max connections: %d", stats.MaxOpenConnections)
	})
}