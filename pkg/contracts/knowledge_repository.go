package contracts

// DirStructure represents the hierarchical structure of directories and files
type DirStructure map[string]DirStructure

// KnowledgeRepository defines the interface for knowledge storage operations
type KnowledgeRepository interface {
	// List returns a json representation of the directory and file structure
	List() (DirStructure, error)

	// Write knowledge to the filesystem
	Write(path string, content string) error

	// Read knowledge from the filesystem
	Read(path string) (string, error)

	// Delete knowledge from the filesystem
	Delete(path string) error
}
