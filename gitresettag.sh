export tagname="v6.0.0"
git push origin :$tagname
git push github :$tagname
git tag --delete $tagname