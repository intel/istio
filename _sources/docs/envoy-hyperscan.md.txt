# Hyperscan

- [Matcher v3 API reference](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/matching/input_matchers/hyperscan/v3alpha/hyperscan.proto)
- [Regex engine v3 API reference](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/regex_engines/hyperscan/v3alpha/hyperscan.proto.html)

[Hyperscan](https://github.com/intel/hyperscan) is a high-performance multiple regex matching library, which uses
hybrid automata techniques to allow simultaneous matching of large numbers of regular expressions and for the matching
of regular expressions across streams of data. Hyperscan supports the
[pattern syntax](https://intel.github.io/hyperscan/dev-reference/compilation.html#pattern-support) used by PCRE.

Hyperscan is only valid in the
[contrib image](https://www.envoyproxy.io/docs/envoy/latest/start/install#install-contrib).

Hyperscan can be used as a matcher of
[generic matching](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/matching/matching), or
enabled as a regex engine globally.

## As a matcher of generic matching

Generic matching has been implemented in a few of components and extensions in Envoy, including
[filter chain matcher](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener.proto.html#envoy-v3-api-field-config-listener-v3-listener-filter-chain-matcher),
[route matcher](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto.html#envoy-v3-api-field-config-route-v3-virtualhost-matcher) and
[RBAC matcher](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/rbac/v3/rbac.proto.html#envoy-v3-api-field-extensions-filters-http-rbac-v3-rbac-matcher).
Hyperscan matcher can be used in generic matcher as a custom matcher in the following structure:

```yaml
custom_match:
  name: hyperscan
  typed_config:
    "@type": type.googleapis.com/envoy.extensions.matching.input_matchers.hyperscan.v3alpha.Hyperscan
    regexes:
    - regex: allowed.*path
```

The behavior of regex matching in Hyperscan matchers can be configured, please refer to the
[API reference](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/matching/input_matchers/hyperscan/v3alpha/hyperscan.proto.html#envoy-v3-api-msg-extensions-matching-input-matchers-hyperscan-v3alpha-hyperscan-regex).

Hyperscan matcher also supports multiple pattern matching which allows matches to be reported for several patterns
simultaneously. Multiple pattern matching can be turned on in the following structure:

```yaml
custom_match:
  name: hyperscan
  typed_config:
    "@type": type.googleapis.com/envoy.extensions.matching.input_matchers.hyperscan.v3alpha.Hyperscan
    # The following multiple patterns match input including allowed.*path and excluding
    # den(y|ied). E.g., the path /allowed/path will be matched, while the path
    # /allowed/denied/path will not be matched.
    regexes:
    - regex: allowed.*path
      id: 1
      quiet: true
    - regex: den(y|ied)
      id: 2
      quiet: true
    - regex: 1 & !2
      combination: true
```

## As a regex engine

Hyperscan regex engine acts in the similar behavior with the default regex engine Google RE2 like it turns on UTF-8
support by default. Hyperscan regex engine can be easily configured with the following configuration.

```yaml
default_regex_engine:
  name: envoy.regex_engines.hyperscan
  typed_config:
    '@type': type.googleapis.com/envoy.extensions.regex_engines.hyperscan.v3alpha.Hyperscan
```
