# 1list-go

A minimal command-line tool for managing task lists stored in `.1list` files.

## Usage

- `./1list set-folder <path>`: Set the folder containing your `.1list` files.
- `./1list`: List tasks and interactively add or toggle them.
- `./1list done <number>`: Toggle a task as done/undone.
- `./1list help`: Show usage info.

## Example

```sh
./1list set-folder ~/Tasks
./1list
./1list done 2
```

## Task List Format

Task lists are stored as JSON files with a `.1list` extension.

## Building

```sh
go build -o 1list main.go
```

## Notes

- The config file is saved in your home directory as `.task-cli-config.json`.
- Only `.1list` files in the configured folder are recognized as task lists.