# rewit

`rewit` is a simple Go CLI tool designed to update the author and committer information in multiple Git repositories.

## Features

- Reads an input file in YAML format containing user information and a list of Git repository URLs.
- Clones each repository.
- Rewrites the commit history to update the author and committer information.
- Force-pushes the changes back to the original repository.
- Allows overriding the input file, user name, and email via command-line flags.

## Usage

1. **Prepare Input File:**
   Create an input file named `rewit.yml` in the following YAML format:
   ```yaml
   user:
     name: John Doe
     email: john.doe@mail.com
   repos:
     - https://github.com/user/repo1.git
     - https://github.com/user/repo2.git
   ```

2. **Run the Tool:**
   Execute the tool without an input file as an argument:
   ```sh
   rewit 
   ```
   There should be a `rewit.yml` file in the dir you are executing the command.

   To specify a different input file:
   ```sh
   rewit -file=anotherfile.yml
   ```

   To override the user name and email via flags:
   ```sh
   rewit -name="Jane Doe" -email="john.doe@mail.com"
   ```

   Combination of different input file and overriding user name and email:
   ```sh
   rewit -file=anotherfile.yml -name="Jane Doe" -email="john.doe@mail.com"
   ```

## Example

To illustrate how to use the tool, let's say we have an input file `rewit.yml` with the following contents:

```yaml
user:
  name: John Doe
  email: john.doe@mail.com
repos:
  - https://github.com/user/repo1.git
  - https://github.com/user/repo2.git
```

Running the tool with this input file will clone the repositories listed and update the author and committer information to "John Doe" and "john.doe@mail.com".

## Notes

- Git should be installed and accessible in your PATH.
- Make sure you have permission to push changes to the repositories listed in the input file.
