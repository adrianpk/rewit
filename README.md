# rewit

`rewit` is a simple Go CLI tool designed to update the author and committer information in multiple Git repositories.

## Features

- Reads an input file containing the name, email, and list of Git repository URLs.
- Clones each repository.
- Rewrites the commit history to update the author and committer information.
- Force-pushes the changes back to the original repository.

## Usage

1. **Prepare Input File:**
   Create an input file named `input.txt` in the following format:
   ```
   John Doe
   john.doe@mail.com
   https://github.com/user/repo1.git
   https://github.com/user/repo2.git
   ```

4. **Run the Tool:**
   Execute the tool with the input file as an argument:
   ```
   rewit input.txt
   ```

## Example

To illustrate how to use the tool, let's say we have an input file `input.txt` like the previous one.

Running the tool with this input file will clone the repositories listed and update the author and committer information to "John Doe" and "john.doe@mail.com".

