(function () {
  // Update this when a new release ships.
  const LATEST_RELEASE = 'v5.0.0';

  function semverLte(a, b) {
    const parse = v => v.replace(/^v/, '').split('.').map(Number);
    const [aMaj, aMin, aPat] = parse(a);
    const [bMaj, bMin, bPat] = parse(b);
    if (aMaj !== bMaj) return aMaj < bMaj;
    if (aMin !== bMin) return aMin < bMin;
    return aPat <= bPat;
  }

  function renderLink(el, tag) {
    el.outerHTML = `Added in gomplate <a href="https://github.com/hairyhenderson/gomplate/releases/tag/${tag}">${tag}</a>`;
  }

  function renderUnreleased(el, tag) {
    el.outerHTML = `Not yet released &mdash; coming in ${tag}`;
  }

  const spans = document.querySelectorAll('span.release-check');
  if (!spans.length) return;

  // Separate known-released tags from ones that need an API check.
  const toCheck = new Set();
  spans.forEach(el => {
    if (!semverLte(el.dataset.tag, LATEST_RELEASE)) {
      toCheck.add(el.dataset.tag);
    }
  });

  // Fetch only the unknown tags, then update all spans.
  // results: 'released' | 'unreleased' | undefined (leave original on error)
  const results = {};
  const checks = Array.from(toCheck).map(tag =>
    fetch(`https://api.github.com/repos/hairyhenderson/gomplate/releases/tags/${encodeURIComponent(tag)}`)
      .then(res => { results[tag] = res.ok ? 'released' : res.status === 404 ? 'unreleased' : undefined; })
      .catch(() => { results[tag] = undefined; })
  );

  Promise.all(checks).then(() => {
    spans.forEach(el => {
      const tag = el.dataset.tag;
      if (semverLte(tag, LATEST_RELEASE) || results[tag] === 'released') {
        renderLink(el, tag);
      } else if (results[tag] === 'unreleased') {
        renderUnreleased(el, tag);
      }
      // undefined (rate-limited, network error, etc.) - leave original text
    });
  });
})();
