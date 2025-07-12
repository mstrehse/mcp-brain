package knowledge

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestSqliteRepo(t *testing.T) (*SqliteRepository, string) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	repo, err := NewSqliteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test repository: %v", err)
	}
	return repo, dbPath
}

func TestSqliteRepository_WriteReadDelete(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	project := "testproj"
	path := "foo/bar.md"
	content := "hello world"

	// Write
	if err := repo.Write(project, path, content); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read
	read, err := repo.Read(project, path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if read != content {
		t.Errorf("Read content mismatch: got %q, want %q", read, content)
	}

	// Delete
	if err := repo.Delete(project, path); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = repo.Read(project, path)
	if err == nil {
		t.Error("Expected error reading deleted file, got nil")
	}
}

func TestSqliteRepository_List_Empty(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	project := "emptyproj"
	// No files written
	list, err := repo.List(project)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("Expected empty DirStructure, got: %+v", list)
	}
}

func TestSqliteRepository_List_Flat(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	project := "flatproj"
	_ = repo.Write(project, "a.md", "A")
	_ = repo.Write(project, "b.md", "B")
	list, err := repo.List(project)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if _, ok := list["a.md"]; !ok {
		t.Error("Expected a.md in DirStructure")
	}
	if _, ok := list["b.md"]; !ok {
		t.Error("Expected b.md in DirStructure")
	}
	// Files should have nil values
	if list["a.md"] != nil {
		t.Error("Expected a.md to have nil value")
	}
	if list["b.md"] != nil {
		t.Error("Expected b.md to have nil value")
	}
}

func TestSqliteRepository_List_Hierarchical(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	project := "hierproj"
	_ = repo.Write(project, "parent/parent.md", "parent content")
	_ = repo.Write(project, "parent/child/child.md", "child content")
	list, err := repo.List(project)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	parent, ok := list["parent"]
	if !ok {
		t.Fatal("Expected 'parent' directory in DirStructure")
	}

	// Since DirStructure is map[string]DirStructure, we can access it directly
	if _, ok := parent["parent.md"]; !ok {
		t.Error("Expected parent.md in parent directory")
	}

	child, ok := parent["child"]
	if !ok {
		t.Fatal("Expected 'child' directory in parent")
	}

	if _, ok := child["child.md"]; !ok {
		t.Error("Expected child.md in child directory")
	}
}

func TestSqliteRepository_ReadDelete_NonExistent(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	project := "nope"
	_, err := repo.Read(project, "missing.md")
	if err == nil {
		t.Error("Expected error reading non-existent file, got nil")
	}
	if err := repo.Delete(project, "missing.md"); err == nil {
		t.Error("Expected error deleting non-existent file, got nil")
	}
}

func TestSqliteRepository_Write_UpdateExisting(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	project := "updateproj"
	path := "test.md"
	content1 := "original content"
	content2 := "updated content"

	// Write original
	if err := repo.Write(project, path, content1); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read original
	read, err := repo.Read(project, path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if read != content1 {
		t.Errorf("Read content mismatch: got %q, want %q", read, content1)
	}

	// Update
	if err := repo.Write(project, path, content2); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Read updated
	read, err = repo.Read(project, path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if read != content2 {
		t.Errorf("Read content mismatch: got %q, want %q", read, content2)
	}
}

func TestSqliteRepository_Delete_Directory(t *testing.T) {
	repo, _ := setupTestSqliteRepo(t)
	defer func() { _ = repo.Close() }()

	project := "delproj"
	// Create files in a directory
	_ = repo.Write(project, "dir/file1.md", "content1")
	_ = repo.Write(project, "dir/file2.md", "content2")
	_ = repo.Write(project, "dir/subdir/file3.md", "content3")

	// Verify files exist
	if _, err := repo.Read(project, "dir/file1.md"); err != nil {
		t.Fatalf("Expected file1.md to exist: %v", err)
	}
	if _, err := repo.Read(project, "dir/subdir/file3.md"); err != nil {
		t.Fatalf("Expected file3.md to exist: %v", err)
	}

	// Delete directory
	if err := repo.Delete(project, "dir"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify all files in directory are gone
	if _, err := repo.Read(project, "dir/file1.md"); err == nil {
		t.Error("Expected dir/file1.md to be deleted")
	}
	if _, err := repo.Read(project, "dir/file2.md"); err == nil {
		t.Error("Expected dir/file2.md to be deleted")
	}
	if _, err := repo.Read(project, "dir/subdir/file3.md"); err == nil {
		t.Error("Expected dir/subdir/file3.md to be deleted")
	}
}

func TestSqliteRepository_DatabasePersistence(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "persistence.db")

	// Create repo and write data
	repo1, err := NewSqliteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	project := "persistproj"
	path := "persistent.md"
	content := "persistent content"

	if err := repo1.Write(project, path, content); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if err := repo1.Close(); err != nil {
		t.Fatalf("Failed to close repository: %v", err)
	}

	// Verify database file exists
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("Database file should exist: %v", err)
	}

	// Create new repo instance with same database
	repo2, err := NewSqliteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create second repository: %v", err)
	}
	defer func() { _ = repo2.Close() }()

	// Read data with new repo instance
	read, err := repo2.Read(project, path)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if read != content {
		t.Errorf("Read content mismatch: got %q, want %q", read, content)
	}
}
