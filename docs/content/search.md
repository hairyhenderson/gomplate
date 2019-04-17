---
title: "Search Results"
sitemap:
  priority : 0.1
---

## Results for _"<span id="search-string"></span>"_

<p>
<em>Showing <span id="search-results-length"></span> results...</em>
</p>

<div>
<div id="search-results">
</div>
</div>

<!-- this template is sucked in by search.js and appended to the search-results div above. So editing here will adjust style -->
<script id="search-result-template" type="text/x-js-template">

<h3><a href="${link}">${title}</a></h3>
<cite><a href="${ link }">${link}</a></cite>
<div id="summary-${key}">
  <div class="search-result">
  ${snippet}&hellip;
  </div>
${ isset tags }<p>Tags: ${tags}</p>${ end }
${ isset categories }<p>Categories: ${categories}</p>${ end }

</div>
</script>
