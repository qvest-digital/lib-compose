# lib-ui-service: composition

This library contains code for the composition of html pages out of multiple html page containing fragments.

## Key Concept
Every Service deliveres a functional html user inteface in form of complete html pages in the way, that a service can be developed and tested
by it's own. A UI-Service can request multimple pages from different services and compose them to one html page. To support this, the html pages
from the services contain a special html vocabular.

## Composition Process
### Loading order

## HTML Composition Vocabulary

### Attribute `uic-remove`
An UI-Service has to remove the element marked with this attribute and all its subelements.

Where: Everywhere (head, body, within framents)

Example:
```
<link uic-remove rel="stylesheet" type="text/css" href="testing.css"/>
```

### Script type `text/uic-meta`
A html page may contain a script of type `text/uic-meta`, with a JSON object as content.
The UI-Service has to add the contents of the JSON object to its global meta data object.

Where: head

Example:
```
<script type="text/uic-meta">
  {
   "foo": "bar",
   "boo": "bazz",
   "categories": ["animal", "human"]
  }
</script>
```
