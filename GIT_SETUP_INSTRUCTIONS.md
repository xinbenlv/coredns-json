# Git Setup Instructions

Follow these steps to initialize a git repository and push to GitHub:

## 1. Create a folder and move all files

```bash
# Create the folder
mkdir -p coredns-json

# Move all files to the folder
mv *.go *.sh *.md Dockerfile .gitignore .dockerignore go.* coredns-json/

# Enter the folder
cd coredns-json
```

## 2. Initialize git repository

```bash
# Initialize a new git repository
git init

# Add all files to the repository
git add .

# Make the initial commit
git commit -m "Initial commit of coredns-json plugin"
```

## 3. Create a new repository on GitHub

1. Go to https://github.com/new
2. Enter "coredns-json" as the repository name
3. Add a description: "CoreDNS plugin for JSON API integration"
4. Choose whether the repository should be public or private
5. Do not initialize with README, .gitignore, or license as we already have these files
6. Click "Create repository"

## 4. Push to GitHub

```bash
# Add the GitHub repository as a remote
git remote add origin https://github.com/xinbenlv/coredns-json.git

# Push to GitHub
git push -u origin main
```

If your default branch is named "master" instead of "main", use:

```bash
git push -u origin master
```

## 5. Verify

Visit https://github.com/xinbenlv/coredns-json to make sure everything has been pushed correctly. 