# Execution Guide for Setting Up coredns-json Repository

## Automated Setup (Recommended)

I've created a shell script that automates most of the process for you. Follow these steps:

1. Make the script executable:
   ```bash
   chmod +x setup_repo.sh
   ```

2. Run the script:
   ```bash
   ./setup_repo.sh
   ```

3. The script will:
   - Create a new directory called `coredns-json`
   - Copy all relevant files to the new directory
   - Initialize a git repository
   - Make the initial commit

4. After the script finishes, you need to:
   - Create a new repository on GitHub (https://github.com/new)
   - Connect your local repository to GitHub
   - Push your code

## Manual Setup (Alternative)

If you prefer to do this process manually, follow these steps:

### 1. Create and setup the directory

```bash
# Create directory
mkdir -p coredns-json

# Copy files (safer than moving)
cp -v *.go *.sh *.md Dockerfile .gitignore .dockerignore go.* coredns-json/

# Remove setup files from destination
rm -f coredns-json/setup_repo.sh coredns-json/EXECUTION_GUIDE.md

# Enter directory
cd coredns-json
```

### 2. Initialize git repository

```bash
# Initialize git
git init

# Add all files
git add .

# Commit
git commit -m "Initial commit of coredns-json plugin"
```

### 3. Create repository on GitHub

1. Go to https://github.com/new
2. Repository name: `coredns-json`
3. Description: `CoreDNS plugin for JSON API integration`
4. Make it Public or Private based on your preference
5. Do NOT initialize with README, .gitignore, or license files
6. Click "Create repository"

### 4. Connect and push to GitHub

```bash
# Add GitHub remote
git remote add origin https://github.com/xinbenlv/coredns-json.git

# Push to GitHub (if your default branch is 'main')
git push -u origin main

# OR if your default branch is 'master'
git push -u origin master
```

## Verify Repository

After pushing to GitHub, visit https://github.com/xinbenlv/coredns-json to ensure everything was uploaded correctly. 