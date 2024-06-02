# rewit

`rewit` is a simple Go CLI tool designed to update the author and committer information in multiple Git repositories.

## Features

- Reads an input file in YAML format containing user information and a list of Git repository URLs.
- Generates the input file automatically by fetching repositories from GitHub based on specified criteria.
- Clones each repository.
- Rewrites the commit history to update the author and committer information.
- Force-pushes the changes back to the original repository.
- Allows overriding the input file, user name, and email via command-line flags.

## Prerequisites

- Git should be installed and accessible in your PATH.
- An environment variable containing your GitHub token should be set up. The default variable is `GITHUB_TOKEN`, but this can be overridden with a command-line flag.
- SSH key authentication should be active in the terminal where you run the tool.

## Usage

The tool works in two main steps:

1. **Generate Configuration File**:
   The initial step generates a configuration file (`rewit.yml`) containing user information and a list of repositories associated with the GitHub token.

   ```shell
   rewit -genyaml -name="John Doe" -email="john.doe@mail.com"
   ```

   This command will:
   - Use the provided name and email for the new commit history.
   - Retrieve the list of repositories associated with the authenticated user.
   - Create a `rewit.yml` file with the user information and repository URLs.

   Additional flags:
   - `-include="substring"`: Only include repositories containing this substring.
   - `-exclude="substring"`: Exclude repositories containing this substring (takes precedence over include).
   - `-token-envar="GITHUB_TOKEN"`: Use a different environment variable for the GitHub token.

2. **Process Repositories**:
   After generating and potentially editing the `rewit.yml` file, execute the tool to rewrite commit history.

   ```shell
   rewit -do
   ```

   This command will:
   - Read the `rewit.yml` file.
   - Clone each repository listed.
   - Rewrite the commit history to update the author and committer information.
   - Force-push the changes back to the original repository.

## Configuration File Format

The configuration file (`rewit.yml`) is generated in the following YAML format:

```yaml
user:
  name: John Doe
  email: john.doe@mail.com
repos:
  - git@github.com:user/repo1
  - git@github.com:user/repo2
```

## Example

1. **Generate Configuration File**:
   ```shell
   rewit -genyaml -name="Jane Doe" -email="jane.doe@mail.com" -include="project" -exclude="archive"
   ```

   This will generate a `rewit.yml` file including repositories that contain "project" in their names and excluding those with "archive".

2. **Edit Configuration File** (Optional):
   Open `rewit.yml` and manually remove any repositories you do not want to process.

3. **Process Repositories**:
   ```shell
   rewit -do
   ```

   This will rewrite the commit history for the repositories listed in `rewit.yml`.

## Notes

- Ensure that the user token has permission to push changes to the repositories listed in the configuration file.
- The tool requires console SSH key authentication for pushing changes.
- Please note that the tool cannot operate on archived repositories. If you need to update an archived repository, you must unarchive it first.

- **Warning**: Rewriting commit history can have significant implications, especially in repositories with multiple collaborators. It is recommended to use this tool only on personal repositories where you are the sole collaborator.

## Testing and Future Development

- The tool has been successfully used for personal purposes but still requires some tests.
