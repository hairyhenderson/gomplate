import csv
import io
import runpy
from pathlib import Path

import pytest


SCRIPT = runpy.run_path(Path(__file__).with_name("benchstat-summary.py"))
parse_benchstat_rows = SCRIPT["parse_benchstat_rows"]
regressions_above_threshold = SCRIPT["regressions_above_threshold"]
render_markdown = SCRIPT["render_markdown"]

BENCHSTAT_CSV = """goos: linux
goarch: amd64
,.tmp/base.txt,,.tmp/head.txt,,,
,sec/op,CI,sec/op,CI,vs base,P
Thing-8,1e-06,± 1%,1.2e-06,± 1%,+20.00%,p=0.008 n=10
Stable-8,2e-06,± 1%,2.01e-06,± 1%,~,p=0.310 n=10
HeadOnly-8,,,3e-06,± 1%
geomean,1e-06,,1.2e-06,,+20.00%,

,.tmp/base.txt,,.tmp/head.txt,,,
,B/op,CI,B/op,CI,vs base,P
Thing-8,200,± 0%,250,± 0%,+25.00%,p=0.008 n=10

,.tmp/base.txt,,.tmp/head.txt,,,
,allocs/op,CI,allocs/op,CI,vs base,P
Thing-8,4,± 0%,3,± 0%,-25.00%,p=0.008 n=10
"""


def test_parse_benchstat_rows_preserves_metric_identity_and_formats_values():
    changes = parse_benchstat_rows(csv.reader(io.StringIO(BENCHSTAT_CSV)))

    assert [(change.name, change.metric, change.base, change.head, change.pct) for change in changes] == [
        ("Thing-8", "sec/op", "1 us/op", "1.2 us/op", 20.0),
        ("Thing-8", "B/op", "200 B/op", "250 B/op", 25.0),
        ("Thing-8", "allocs/op", "4 allocs/op", "3 allocs/op", -25.0),
    ]


def test_render_markdown_emphasizes_only_regressions_above_threshold():
    changes = parse_benchstat_rows(csv.reader(io.StringIO(BENCHSTAT_CSV)))

    rendered = render_markdown(changes, threshold=20)

    assert "| Benchmark | Metric | Base | Head | Change | p-value |" in rendered
    assert "| `Thing-8` | sec/op | 1 us/op | 1.2 us/op | +20.00% | 0.008 |" in rendered
    assert "| `Thing-8` | B/op | 200 B/op | 250 B/op | **+25.00%** | 0.008 |" in rendered
    assert "<summary>1 improvement(s)</summary>" in rendered


def test_regression_gate_checks_each_metric():
    changes = parse_benchstat_rows(csv.reader(io.StringIO(BENCHSTAT_CSV)))

    regressions = regressions_above_threshold(changes, threshold=20)

    assert [(change.metric, change.pct) for change in regressions] == [("B/op", 25.0)]


def test_parse_benchstat_rows_rejects_non_csv_output():
    with pytest.raises(ValueError, match="no benchstat CSV metric tables"):
        parse_benchstat_rows(csv.reader(io.StringIO("BenchmarkThing 1 ns/op\n")))
