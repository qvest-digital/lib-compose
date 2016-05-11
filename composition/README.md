# lib-ui-service: composition

This library contains code for the composition of HTML pages out of multiple HTML page containing fragments.

### Key Concept
Every Service delivers a functional HTML user interface in form of complete HTML pages in the way, that a service can be developed and tested
by it's own. A UI-Service can request multiple pages from different services and compose them to one HTML page. To support this, the HTML pages
from the services contain a special HTML vocabular.

### Composition Process
The composition is done in the following steps:

1. The UI-Service has a `CompositionHandler` in it's handler chain, which answers these which need composition.
2. The `CompositionHandler` has a callback from the UI-Service. This callback gets a `http.Request` object as argument and returns a List of FetchResult.
3. For each request this callback is triggered. So the UI-Service can add a `ContentFetcher` for this request and adds FetchDefinitions for Page using `ContentFetcher.AddFetchJob()`.
4. The ContentFetchter loads the Pages and recursively it's dependencies in parallel. For the actual loading and parsing, it uses the `HtmlContentLoader`.
5. When all `Content` objects are loaded, the `CompositionHandler` merges them together, using `ContentMerge`.

### Merging
The it self is very simple:

- The MetaJSON is calculated by adding all fields of the loaded MetaJSON to one global map.
- All Head fragments are concatenated within the `<head>`.
- The Default Body Fragment of the first Content object is executed and may recursively include other Fragments Content objects.
- All Tail fragments are concatenated at the end of the `<body>`.

### Execution Order
**Attention**: The execution order of the Content Objects is determined by the order in which they are returned from the `ContentFetcher`.
Currently this is only deterministic within the FetchDefinitions added by `ContentFetcher.AddFetchJob()`. The recursive dependencies are loaded from them are in a random order.
This may cause not deterministic behaviour, if the contain equal named fragments or provide the same MetaJSON attributes.

### Caching
At a later point, the ContentFetcher may provide Caching of FetchDefinitions.



## HTML Composition Vocabulary

### Attribute `uic-remove`
An UI-Service has to remove the element marked with this attribute and all its subelements.
Be careful to have a correct open and closing structure in the HTML. The standard selfclosing tags are
allowed, e.g. both are working `<br>` and `<br/>`, but if there is a structure error with e.g. a div,
`uic-remove` may lead to strage behaviour.

Example:

```HTML
<link uic-remove rel="stylesheet" type="text/css" href="testing.css"/>
```

Where: Everywhere (head, body, within fragments)

### Script type `text/uic-meta`
A HTML page may contain a script of type `text/uic-meta`, with a JSON object as content.
The UI-Service has to add the contents of the JSON object to its global meta data object.

Example:

```HTML
<script type="text/uic-meta">
  {
   "foo": "bar",
   "boo": "bazz",
   "categories": ["animal", "human"]
  }
</script>
```

Where: head


### Fragments
The UI-Service interpretes an HTML page as a set of fragments. All those fragments are optional.

- One __Head Fragment__, identified by the child elements of the HTML `<head>` tag.
- One __Body Default Fragement__, identified by the child elements of the `<body>` tag or by a `uic-fragment` without a name attribute.
- Multiple __Named Body Fragments__, identified by `uic-fragment` tag within the body.
- One __Tail Fragment__, identified by the `uic-tail` tag.

#### Head-Fragment
The complete contents of the head is interpreted as the head fragment. The elements marked with `uic-remove`
and the `uic-meta` Script are not cleaned out of the head fragment. If the head framents only contains whitespace,
it is interpreted as not existing.

Example: The Head Fragment contains `<title>The Title</title>`

```HTML
<head>
  <title>The Title</title>
  <link uic-remove rel="stylesheet" type="text/css" href="special.css"/>
  <script type="text/uic-meta">
    {}
  </script>
</head>
```

#### Body Default Fragment 
All other elements fragments and those elements, marked with `uic-remove`, are removed from the body
and the remaining fragment is taken as Body Default Fragment. The Body Default Fragment is just a fragment with
the empty name (""). If there is a `uic-fragment` tag without the name in the body, this overwrites the default fragment.

Example: The Default Fragment contains `<h1>Hello World</h1>`

```HTML
<body>
    Hello World
    <ul uic-remove>
      <!-- A Navigation for testing -->
    </ul>
    <uic-fragment name="headline">
      <h1>This is a headline</h1>
    </uic-fragment>
</body>
```

The complete contents of the body is interpreted as the head fragment. The elements marked with `uic-remove`
and the `uic-meta` script are not cleaned out of the head fragment. If the head fragments only contains whitespace,
it is interpreted as not existing.

Example: The Default Fragment contains `<h1>This is the default</h1>`

```HTML
<body>
    <h1>Hello World</h1>
    <uic-fragment>
      <h1>This is the default</h1>
    </uic-fragment>
</body>
```

#### Element `uic-fragment`
The body of an HTML page may contain multiple `uic-fragment` tags. Which contain the fragments for the page.
All content within the tag is taken as fragment content. Nested Fragment tags are not allowed.

The Fragment Tag may have a `name` attribute, for named the fragment. If no attribute is given, or the name is empty,
the Body Default Fragment is overwritten by this fragment.

Example: Contains two fragments *headline* and *w*

```HTML
<body>
  <uic-fragment name="headline">
    <h1>This is a headline</h1>
  </uic-fragment>
  <uic-fragment name="w">
    Bli Bla blub
    <div uic-remove>
       Some element for testing
    </div>
  </uic-fragment>
</body>
```

Where: body

### Templating
All fragments (except the Head Fragment) may contain minimal templating directives which has to be resolved by the UI-Service.
There are two forms of includes and a syntax for variable replacement.

#### Variables
The UI-Service has to replace Variables by the corresponding path out of the global meta data.
If the variable name contains a '.', at first, it is tried to match the full path as one string, after that,
it is tried to travese a tree of maps.

Example:
```
§[ foo ]§
```
or
```
§[ foo.bar ]§ // tried to match MetaJSON['foo.bar'] and than MetaJSON['foo']['bar']
```

#### Predefined Variables
There are some predefined variables, constructed out of the request.
```
{'request': {
    'base_url': 'http://example.com/' // the base url of the service, calculated out of the request, e.g.
    'params: {..} // a map with the GET Query parameters of the request.
  }
}
```

#### Preloaded Includes 
On an unspecified include, the UI-Service has to load replace the include by a previously loaded fragment.

Example: Will be replaced by the Default Body Fragment of *example.com/foo*.

```
§[> example.com/foo]§
```

Example: Will be replaced by the *content* fragment of *example.com/foo*.

```
§[> example.com/foo#content]§
```

Example: Will be replaced by the *content* fragment of any random choosen page.

```
§[> #content]§
```

#### Loaded Includes 
On a specified include, the UI-Service has to load the referenced page and has to replace the include with the referenced fragment.
Within the src attribute, there are also variable replacements possible.

Example: Will be replaced by the Default Body Fragment of *http://example.com/foo*.

```
  <uic-include src="example.com/foo"/>
```

Example: Will be replaced by the *content* fragment of *http://example.com/foo*. If it times out after 42 seconds, no error is returned.

```
  <uic-include src="example.com/foo#content" timeout="42000" required="false"/>
```

