#!/usr/bin/env bash
# Tags a release of main and pushes the tag, which starts .github/workflows/release.yml
# (GoReleaser, signed GitHub release, then the Terraform Registry picks it up).
set -euo pipefail

REMOTE="origin"
RELEASE_BRANCH="main"
CHANGELOG="CHANGELOG.md"

usage() {
  cat >&2 <<'EOF'
usage: task release [-- [vX.Y.Z | vX.Y.Z-rc.N] [--yes]]
  no version   tag the version of the newest section in CHANGELOG.md
  vX.Y.Z       a final release: CHANGELOG.md must already have its section
  vX.Y.Z-rc.N  a release candidate: no CHANGELOG section needed
  --yes        skip the confirmation prompt
EOF
  exit 2
}

new=""
yes=false
for arg in "$@"; do
  case "$arg" in
    -y|--yes) yes=true ;;
    -h|--help) usage ;;
    v*) [ -z "$new" ] || usage; new="$arg" ;;
    *) usage ;;
  esac
done

fail() { echo "Error: $*" >&2; exit 1; }

cd "$(git rev-parse --show-toplevel)"

FINAL_RE='^v[0-9]{1,5}\.[0-9]{1,5}\.[0-9]{1,5}$'
RC_RE='^v[0-9]{1,5}\.[0-9]{1,5}\.[0-9]{1,5}-rc\.[0-9]{1,4}$'
is_final() { echo "$1" | grep -Eq "$FINAL_RE"; }
is_rc() { echo "$1" | grep -Eq "$RC_RE"; }
base_of() { echo "${1%%-rc.*}"; }

# Sortable integer, sized by FINAL_RE and RC_RE: a final release ranks above its release candidates
ver_num() {
  local base="${1#v}" rc=9999 major minor patch
  case "$base" in *-rc.*) rc="${base##*-rc.}"; base="${base%%-rc.*}" ;; esac
  IFS='.' read -r major minor patch <<< "$base"
  echo $(( 10#$major * 100000000000000 + 10#$minor * 1000000000 + 10#$patch * 10000 + 10#$rc ))
}

# Tags on the remote, so a local-only tag never decides the next version
remote_tags() {
  local refs
  refs="$(git ls-remote --tags --refs "$REMOTE" 'v*')" || return 1
  printf '%s\n' "$refs" | sed -E 's#^.*refs/tags/##'
}

# Highest tag of the list on stdin that matches the filter; empty when none matches
highest_of() {
  { grep -E "$1" || true; } | while read -r t; do echo "$(ver_num "$t") $t"; done \
    | sort -n | tail -1 | cut -d' ' -f2
}

require_in_sync() {
  git fetch --quiet "$REMOTE" "$RELEASE_BRANCH" || fail "fetching $REMOTE failed (see the git error above)."
  [ "$(git rev-parse HEAD)" = "$(git rev-parse "$REMOTE/$RELEASE_BRANCH")" ] \
    || fail "${1:-local $RELEASE_BRANCH differs from $REMOTE/$RELEASE_BRANCH. Pull or push first.}"
}

SECTION_RE='^## \[[0-9]+\.[0-9]+\.[0-9]+\] - [0-9]{4}-[0-9]{2}-[0-9]{2}$'

# Newest released section heading: "## [1.8.0] - 2026-11-19" gives 1.8.0
newest_changelog_version() {
  { grep -E "$SECTION_RE" "$CHANGELOG" || true; } | head -1 | sed -E 's/^## \[([^]]+)\].*/\1/'
}

has_changelog_section() {
  { grep -E "$SECTION_RE" "$CHANGELOG" || true; } | grep -qF "## [$1] - "
}

# Commit a remote tag points at; main was fetched, so a released commit is already local
remote_tag_commit() {
  local refs sha
  refs="$(git ls-remote "$REMOTE" "refs/tags/$1" "refs/tags/$1^{}")" || return 1
  sha="$(printf '%s\n' "$refs" | awk '/\^\{\}$/ {print $1}' | head -1)"
  [ -n "$sha" ] || sha="$(printf '%s\n' "$refs" | awk 'NR==1 {print $1}')"
  [ -n "$sha" ] && git cat-file -e "$sha^{commit}" 2>/dev/null && echo "$sha"
}

# Fails unless $new is above every release and candidate it must follow; sets $since
check_order() {
  local tags="$1" base latest_rc
  current_final="$(printf '%s\n' "$tags" | highest_of "$FINAL_RE")"
  [ -n "$current_final" ] || fail "no vX.Y.Z tag found on $REMOTE."
  printf '%s\n' "$tags" | grep -qxF "$new" && fail "tag $new already exists on $REMOTE."
  if is_final "$new"; then
    [ "$(ver_num "$new")" -gt "$(ver_num "$current_final")" ] || fail "$new is not above the latest release $current_final."
    since="$current_final"
  else
    base="$(base_of "$new")"
    [ "$(ver_num "$base")" -gt "$(ver_num "$current_final")" ] || fail "$base is not above the latest release $current_final."
    latest_rc="$(printf '%s\n' "$tags" | highest_of "^${base//./\\.}-rc\.[0-9]{1,4}$")"
    [ -z "$latest_rc" ] || [ "$(ver_num "$new")" -gt "$(ver_num "$latest_rc")" ] \
      || fail "$new is not above the latest candidate $latest_rc."
    since="${latest_rc:-$current_final}"
  fi
}

# Lines between "## [Unreleased]" and the next "## " heading that hold an entry
unreleased_entries() {
  awk '/^## \[Unreleased\]/{in_section=1; next} /^## /{in_section=0} in_section && /^- /' "$CHANGELOG"
}

branch="$(git rev-parse --abbrev-ref HEAD)"
[ "$branch" = "$RELEASE_BRANCH" ] || fail "on branch '$branch', releases are cut from '$RELEASE_BRANCH'."
[ -z "$(git status --porcelain)" ] || fail "working tree not clean. Commit or stash first."

echo "==> Fetching $REMOTE"
require_in_sync

tags="$(remote_tags)" || fail "listing the tags of $REMOTE failed (see the git error above)."
current_final="$(printf '%s\n' "$tags" | highest_of "$FINAL_RE")"
[ -n "$current_final" ] || fail "no vX.Y.Z tag found on $REMOTE."

if [ -z "$new" ]; then
  section="$(newest_changelog_version)"
  [ -n "$section" ] || fail "no '## [X.Y.Z] - YYYY-MM-DD' section in $CHANGELOG. Cut one in a PR first (AGENTS.md, Releasing)."
  new="v$section"
  [ "$(ver_num "$new")" -gt "$(ver_num "$current_final")" ] \
    || fail "the newest $CHANGELOG section is $section, not above the latest release $current_final. Cut the next section in a PR first (AGENTS.md, Releasing), or pass a vX.Y.Z-rc.N version."
fi

is_final "$new" || is_rc "$new" || fail "'$new' is neither vX.Y.Z nor vX.Y.Z-rc.N (at most 5 digits per number, 4 for the candidate)."
check_order "$tags"
git rev-parse -q --verify "refs/tags/$new" >/dev/null && fail "tag $new exists locally but not on $REMOTE: delete it with 'git tag -d $new' if it is a leftover."

if is_final "$new"; then
  has_changelog_section "${new#v}" \
    || fail "$CHANGELOG has no '## [${new#v}] - YYYY-MM-DD' section. Cut it in a PR first (AGENTS.md, Releasing)."
  [ -z "$(unreleased_entries)" ] \
    || fail "$CHANGELOG still has entries under [Unreleased]: they would ship in $new without being listed in its section."
fi

since_commit="$(remote_tag_commit "$since")" \
  || fail "the commit of $since is not in the local history. Fetch it ('git fetch $REMOTE tag $since') and re-run."
commits="$(git log --oneline "$since_commit..HEAD")"
[ -n "$commits" ] || fail "no commit since $since, nothing to release."

remote_url="$(git remote get-url "$REMOTE" | sed -E 's#://[^/@]*@#://***@#')"

echo
echo "  latest release : $current_final"
echo "  new tag        : $new"
echo "  remote         : $REMOTE  $remote_url"
echo "  commit         : $(git log -1 --oneline HEAD)"
echo
echo "  $(echo "$commits" | wc -l | tr -d ' ') commit(s) since $since:"
while IFS= read -r line; do echo "    $line"; done <<< "$commits"
echo
echo "  Pushing $new starts .github/workflows/release.yml: GoReleaser builds and signs the"
echo "  binaries and publishes the GitHub release, then the Terraform Registry picks it up."
echo

if [ "$yes" != true ]; then
  [ -t 0 ] || fail "not a TTY, rerun with --yes to skip the confirmation."
  printf '  Create and push tag %s? [y/N] ' "$new"
  read -r answer || answer=""
  case "$answer" in
    y|Y|yes|YES) ;;
    *) echo "Aborted, nothing pushed."; exit 1 ;;
  esac
fi

# main or the tags may have moved while the prompt was open
require_in_sync "$REMOTE/$RELEASE_BRANCH moved meanwhile, nothing pushed. Re-run to release the new HEAD."
tags="$(remote_tags)" || fail "listing the tags of $REMOTE failed, nothing pushed."
checked_since="$since"
check_order "$tags"
[ "$since" = "$checked_since" ] || fail "$since was tagged on $REMOTE meanwhile, nothing pushed. Re-run to check the new order."

git tag -a "$new" -m "$new"
if ! git push "$REMOTE" "refs/tags/$new"; then
  git tag -d "$new" >/dev/null
  fail "pushing $new failed, local tag removed. Nothing was released."
fi

echo
echo "Pushed $new. Release workflow:"
echo "  https://github.com/Qovery/terraform-provider-qovery/actions/workflows/release.yml"
echo "Registry (a few minutes after the workflow):"
echo "  https://registry.terraform.io/providers/Qovery/qovery/latest"
