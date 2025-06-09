# Importer System Conversion

This document describes the conversion of the question importer system from Python to Go, including the creation of a CLI tool for importing questions from gotquestions.online.

## Overview

The importer system fetches question data from gotquestions.online and stores it in the database. It consists of:

1. **Package Lister** - Fetches list of available question packages
2. **Package Parser** - Fetches all questions from a specific package  
3. **Next.js Helper** - Extracts data from Next.js pages
4. **CLI Tool** - Provides command-line interface for import operations

## Components Implemented

### 1. Next.js Helper (`internal/importer/nextjs_helper.go`)
- Extracts Next.js `__NEXT_DATA__` from HTML pages
- Parses JSON data from script tags
- Handles HTTP requests with proper error handling

**Key functions:**
- `ExtractNextProps()` - Extract props from HTTP response
- `ExtractNextPropsFromURL()` - Fetch URL and extract props

### 2. Package Model (`internal/models/package.go`)
- Database operations for question packages
- Package data structure and validation

**Key methods:**
- `BuildPackageFromDict()` - Create package from API data
- `Insert()` - Store package in database
- `GetPackagesByYear()` - Filter packages by year
- `Exists()` - Check if package already exists

### 3. Package Lister (`internal/importer/package_lister.go`)
- Fetches package lists from gotquestions.online pagination
- Processes multiple pages with rate limiting
- Stores package metadata in database

**Key features:**
- Configurable page range processing
- Automatic rate limiting with random delays
- Error handling for individual pages
- Duplicate detection and skipping

### 4. Package Parser (`internal/importer/package_parser.go`)  
- Fetches all questions from a specific package
- Handles both pack/tours and direct tour structures
- Downloads and stores question images
- Optional rewrite mode for updating existing data

**Key features:**
- Support for multiple question data structures
- Image downloading and storage
- Transactional question insertion
- Question validation and error handling

### 5. Extended Question Model (`internal/models/question.go`)
- Enhanced question structure with import fields
- Question building from API response data

**New fields added:**
- `GotQuestionsID` - Original ID from gotquestions.online
- `AuthorID` - Question author reference
- `PackageID` - Package association  
- `Difficulty` - Question difficulty rating
- `IsIncorrect` - Question validity flag

### 6. CLI Tool (`cmd/importer/main.go`)
- Command-line interface for all import operations
- Three main commands with flexible options

## CLI Commands

### 1. List Packages
```bash
./bin/importer -command=list-packages [-first-page=1] [-last-page=337]
```
- Updates the list of packages in the database
- Processes specified page range (default: all pages)
- Skips existing packages automatically

### 2. Import Package
```bash
./bin/importer -command=import-package -package-id=5220 [-rewrite=true]
```
- Fetches questions for a specific package ID
- Optional rewrite mode to update existing questions
- Downloads and stores question images

### 3. Import Year
```bash
./bin/importer -command=import-year -year=2022 [-rewrite=true]
```
- Fetches questions from all packages for a specific year
- Automatically finds relevant packages by end date
- Skips packages that already have questions (unless rewrite=true)
- Processes packages with rate limiting

## Key Improvements Over Python Version

### Better Error Handling
- Proper Go error propagation
- Graceful handling of individual package failures
- Detailed error logging with context

### Enhanced CLI Interface
- Single binary with multiple commands
- Clear help text and examples
- Flexible parameter options

### Performance Optimizations
- Concurrent-safe operations
- Efficient database queries
- Memory-efficient image handling

### Rate Limiting
- Random delay variations to avoid detection
- Configurable timing between requests
- Respectful of source website resources

## Database Schema Requirements

The importer expects these tables:

```sql
-- Packages table
CREATE TABLE packages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gotquestions_id INTEGER UNIQUE,
    title TEXT,
    start_date DATETIME,
    end_date DATETIME,
    questions_count INTEGER
);

-- Enhanced questions table  
CREATE TABLE questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gotquestions_id INTEGER,
    question TEXT,
    answer TEXT,
    accepted_answer TEXT,
    comment TEXT,
    handout_str TEXT,
    source TEXT,
    author_id INTEGER,
    package_id INTEGER,
    difficulty REAL,
    is_incorrect BOOLEAN
);

-- Images table for question attachments
CREATE TABLE images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    question_id INTEGER,
    image_url TEXT,
    data BLOB,
    mime_type TEXT,
    FOREIGN KEY (question_id) REFERENCES questions(id)
);
```

## Usage Examples

### Basic Package Update
```bash
# Update package list from first 100 pages
./bin/importer -command=list-packages -last-page=100
```

### Import Specific Package
```bash
# Import questions from package 5220
./bin/importer -command=import-package -package-id=5220
```

### Import Full Year
```bash  
# Import all 2022 questions
./bin/importer -command=import-year -year=2022

# Force reimport of 2022 questions  
./bin/importer -command=import-year -year=2022 -rewrite=true
```

## Error Handling

The importer handles various error conditions gracefully:

- **Network failures**: Retry logic for transient issues
- **Parsing errors**: Skip malformed data, continue processing
- **Database errors**: Proper transaction handling
- **Rate limiting**: Automatic delays and backoff

## Data Validation

All imported data goes through validation:

- **Required fields**: Question text, answer, package association
- **Data types**: Proper type conversion from JSON
- **Duplicates**: Prevention of duplicate question imports
- **Images**: MIME type detection and size validation

## Compatibility

The Go importer maintains full compatibility with:
- Existing Python-created database schemas
- Question data formats and structures  
- Image storage mechanisms
- Package metadata standards

The converted system is ready for production use and can fully replace the Python importer while providing enhanced functionality and better performance.