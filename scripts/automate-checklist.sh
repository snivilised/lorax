
function auto-check() {
  local owner=$(git config --get remote.origin.url | cut -d '/' -f 4)
  local repo=$(git rev-parse --show-toplevel | xargs basename)

  echo "---> 😎 OWNER: $owner"
  echo "---> 🧰 REPO: $repo"
  echo ""

  update-workflow-names $repo $owner
  if [ $? -ne 0 ]; then
    return 1
  fi

  update-mod-file $repo $owner
  if [ $? -ne 0 ]; then
    return 1
  fi

  update-source-id-variable-in-translate-defs $repo $owner
  if [ $? -ne 0 ]; then
    return 1
  fi

  update-astrolib-in-taskfile $repo $owner
  if [ $? -ne 0 ]; then
    return 1
  fi

  update-astrolib-in-goreleaser $repo $owner
  if [ $? -ne 0 ]; then
    return 1
  fi

  rename-templ-data-id $repo $owner
  if [ $? -ne 0 ]; then
    return 1
  fi

  update-import-statements $repo $owner
  if [ $? -ne 0 ]; then
    return 1
  fi

  update-readme $repo $owner
  if [ $? -ne 0 ]; then
    return 1
  fi

  rename-language-files $repo $owner
  if [ $? -ne 0 ]; then
    return 1
  fi

  reset-version
  if [ $? -ne 0 ]; then
    return 1
  fi

  touch ./.env
  echo "✔️ done"
  return 0
}

# the sed -i option edits the file in place, overwriting the original file
#
function update-all-generic() {
  local title=$1
  local repo=$2
  local owner=$3
  local folder=$4
  local name=$5
  local target=$6
  local replacement=$7

  echo "  🎯 --->        title: $title"
  echo "  ✅ ---> file pattern: $name"
  echo "  ✅ --->       folder: $folder"
  echo "  ✅ --->       target: $target"
  echo "  ✅ --->  replacement: $replacement"

  find $folder -name "$name" -type f -print -exec sed -i "s/${target}/${replacement}/g" {} +

  if [ $? -ne 0 ]; then
    echo "!!! ⛔ Aborted! update-all-generic failed for $owner/$repo:"
    return 1
  fi

  echo "  ✔️ --->  DONE"
  echo ""
  return 0
}

function update-mod-file() {
  local repo=$1
  local owner=$2
  local folder=./
  local file_pattern=go.mod
  local target="module github.com\/snivilised\/astrolib"
  local replacement="module github.com\/$owner\/$repo"
  update-all-generic "update-mod-file" $repo $owner $folder $file_pattern "$target" "$replacement"
}

function update-source-id-variable-in-translate-defs() {
  local repo=$1
  local owner=$2
  local folder=./i18n/
  local file_pattern=translate-defs.go
  local target="AstrolibSourceID"
  local tc_repo="$(echo ${repo:0:1} | tr '[:lower:]' '[:upper:]')${repo:1}"
  local replacement="${tc_repo}SourceID"
  update-all-generic "update-source-id-variable-in-translate-defs" $repo $owner $folder $file_pattern "$target" "$replacement"
}

function update-astrolib-in-taskfile() {
  local repo=$1
  local owner=$2
  local folder=./
  local file_pattern=Taskfile.yml
  local target=astrolib
  local replacement=$repo
  update-all-generic "update-astrolib-in-taskfile" $repo $owner $folder $file_pattern "$target" "$replacement"
}

function update-workflow-names() {
  local repo=$1
  local owner=$2
  local folder=.github/workflows
  local file_pattern=*.yml
  local target="name: astrolib"
  local tc_repo="$(echo ${repo:0:1} | tr '[:lower:]' '[:upper:]')${repo:1}"
  local replacement="name: $tc_repo"
  update-all-generic "💥 update-workflow-names" $repo $owner $folder $file_pattern "$target" $replacement
}

function update-astrolib-in-goreleaser() {
  local repo=$1
  local owner=$2
  local folder=./
  local file_pattern=.goreleaser.yaml
  local target=astrolib
  local replacement=$repo
  update-all-generic "update-astrolib-in-goreleaser" $repo $owner $folder $file_pattern "$target" $replacement
}

function rename-templ-data-id() {
  local repo=$1
  local owner=$2
  local folder=./
  local file_pattern=*.go
  local target="astrolibTemplData"
  local replacement="${repo}TemplData"
  update-all-generic "rename-templ-data-id" $repo $owner $folder $file_pattern "$target" "$replacement"
}

function update-readme() {
  local repo=$1
  local owner=$2
  local folder=./
  local file_pattern=README.md
  local target="astrolib: "
  local replacement="${repo}: "

  update-all-generic "update-readme(astrolib:)" $repo $owner $folder $file_pattern "$target" "$replacement"
  if [ $? -ne 0 ]; then
    return 1
  fi

  target="snivilised\/astrolib"
  replacement="$owner\/$repo"
  update-all-generic "update-readme(snivilised/astrolib)" $repo $owner $folder $file_pattern "$target" "$replacement"
  if [ $? -ne 0 ]; then
    return 1
  fi

  target="astrolib Continuous Integration"
  tc_repo="$(echo ${repo:0:1} | tr '[:lower:]' '[:upper:]')${repo:1}"
  replacement="$tc_repo Continuous Integration"
  update-all-generic "update-readme(astrolib Continuous Integration)" $repo $owner $folder $file_pattern "$target" "$replacement"
  if [ $? -ne 0 ]; then
    return 1
  fi

  return 0
}

function update-import-statements() {
  local repo=$1
  local owner=$2
  local folder=./
  local file_pattern=*.go
  local target="snivilised\/astrolib"
  local replacement="$owner\/$repo"
  update-all-generic "update-import-statements" $repo $owner $folder $file_pattern "$target" "$replacement"
}

function rename-language-files() {
  local repo=$1
  find . -name 'astrolib*.json' -type f -print0 |
  while IFS= read -r -d '' file; do
    mv "$file" "$(dirname "$file")/$(basename "$file" | sed "s/^astrolib/$repo/")"
  done
  return $?
}

function reset-version() {
  echo "v0.1.0" > ./VERSION
  return 0
}
