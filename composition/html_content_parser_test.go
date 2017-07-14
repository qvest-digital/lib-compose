package composition

import (
	"bytes"
	"fmt"
	_ "regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

var productUiGeneratedHtml = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>navigationservice</title>
    <!-- START Include legacy styles - emulate integration -->

    <!-- END Include legacy styles -->
    <link rel="stylesheet" href="/navigationservice/stylesheets/main-ffc9b54a22.css">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="canonical" href="/baumarkt/suche">
    <script>
        // Define global SCRIPTS variable and
        // global loadScript() method to loading scripts
        // async but in order.
        // Each module register it's javascript by calling
        // this method:
        //
        // loadScript('/navigationservice/components/molecules/teaser/teaser.js');
        //
        SCRIPTS = ['/navigationservice/javascripts/vendor/jquery-8101d596b2.min.js', '/navigationservice/javascripts/main-680c12b0b1.js'];
        isLegacy = function() {
            return typeof Object.assign === 'function' ? false : true;
        };

        loadScript = function(script, legacyOnly) {
            for(var i=0; i < SCRIPTS.length; i++) if(SCRIPTS[i] === script) return false;
            if((legacyOnly && isLegacy()) || (!legacyOnly)) {
                SCRIPTS.push(script);
            }
        };
    </script>

    <!-- fonts.com - Async Font Loading -->
    <script type="text/javascript">
        (function() {
            var fontsComPath = '//fast.fonts.net/jsapi/0d47d266-cd84-4ef7-adb8-5a44ad7011ef.js',
                    fontsComJS = document.createElement('script');
            fontsComJS.type = 'text/javascript';
            fontsComJS.async = true;
            fontsComJS.src = fontsComPath;
            var head = document.getElementsByTagName("head")[0];
            head.appendChild(fontsComJS);
        })();
    </script><meta charset="utf-8">
    <!--
        This website is powered by TYPO3 - inspiring people to share!
        TYPO3 is a free open source Content Management Framework initially created by Kasper Skaarhoj and licensed under GNU/GPL.
        TYPO3 is copyright 1998-2016 of Kasper Skaarhoj. Extensions are copyright of their respective owners.
        Information and contribution at http://typo3.org/
    -->
    <base href="/">
    <meta name="generator" content="TYPO3 CMS">
    <meta name="content-language" content="de">
    <link rel="stylesheet" type="text/css" href="typo3temp/compressor/merged-d0ed097d2e70237fa36186d357e1268f-4e221af468cdd1d3a44789532134127c.css?1476243484" media="all">
    <script src="typo3temp/compressor/merged-f6a1f7cc0a094340acf2489928881fc7-956c525da07a115d310e68d089faa490.js?1476243484" type="text/javascript"></script>
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta http-equiv="cleartype" content="on">
    <meta name="mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="black">
    <link rel="canonical" href="/navigationservice">
    <link rel="stylesheet" href="typo3conf/ext/bra_projectfiles_toom/Resources/Public/website/css/main.css">

    <link rel="shortcut icon" href="favicon.ico" type="image/ico" />
    <link rel="icon" href="favicon.ico" type="image/ico" />

    <!-- picturefill:start -->
    <script src="typo3conf/ext/bra_projectfiles_toom/Resources/Public/website/js/libs/vendor/picturefill/picturefill.min.js" async></script>
    <!-- picturefill:end --><link href="http://www.toom-baumarkt.de/navigation/" rel="canonical"><meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta http-equiv="cleartype" content="on">
    <meta name="mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-status-bar-style" content="black">

    <link type="text/css" rel="stylesheet" href="//fast.fonts.net/cssapi/0d47d266-cd84-4ef7-adb8-5a44ad7011ef.css"/>

    <link rel="stylesheet" href="/rebrush/assets/typo/stylesheets/main-61a49a7baa.css">

    <!-- picturefill:start -->
    <script src="/rebrush/assets/typo/javascripts/picturefill-f350acdff4.min.js" async></script>
    <!-- picturefill:end --><meta charset="utf-8">
    <title>navigationservice</title>


    <!-- START Include legacy styles - emulate integration -->

    <!-- END Include legacy styles -->

    <link rel="stylesheet" href="/navigationservice/stylesheets/main-ffc9b54a22.css">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">

    <script>
        // Define global SCRIPTS variable and
        // global loadScript() method to loading scripts
        // async but in order.
        // Each module register it's javascript by calling
        // this method:
        //
        // loadScript('/navigationservice/components/molecules/teaser/teaser.js');
        //
        SCRIPTS = ['/navigationservice/javascripts/vendor/jquery-8101d596b2.min.js', '/navigationservice/javascripts/main-680c12b0b1.js'];
        isLegacy = function() {
            return typeof Object.assign === 'function' ? false : true;
        };

        loadScript = function(script, legacyOnly) {
            for(var i=0; i < SCRIPTS.length; i++) if(SCRIPTS[i] === script) return false;
            if((legacyOnly && isLegacy()) || (!legacyOnly)) {
                SCRIPTS.push(script);
            }
        };
    </script>

    <!-- fonts.com - Async Font Loading -->
    <script type="text/javascript">
        (function() {
            var fontsComPath = '//fast.fonts.net/jsapi/0d47d266-cd84-4ef7-adb8-5a44ad7011ef.js',
                    fontsComJS = document.createElement('script');
            fontsComJS.type = 'text/javascript';
            fontsComJS.async = true;
            fontsComJS.src = fontsComPath;
            var head = document.getElementsByTagName("head")[0];
            head.appendChild(fontsComJS);
        })();
    </script><meta charset="utf-8">
    <title>Suchergebnis | toom Baumarkt</title>

    <meta name="viewport" content="width=device-width, initial-scale=3.0">

    <link rel="canonical" href="/baumarkt/suche">

    <!-- START Include legacy styles - emulate integration -->

    <!-- END Include legacy styles -->

    <link rel="stylesheet" href="/searchservice/stylesheets/main-ffc9b54a22.css">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">

    <script>
        // Define global SCRIPTS variable and
        // global loadScript() method to loading scripts
        // async but in order.
        // Each module register it's javascript by calling
        // this method:
        //
        // loadScript('/searchservice/components/molecules/teaser/teaser.js');
        //
        SCRIPTS = ['/searchservice/javascripts/vendor/jquery-8101d596b2.min.js', '/searchservice/javascripts/main-680c12b0b1.js'];
        isLegacy = function() {
            return typeof Object.assign === 'function' ? false : true;
        };

        loadScript = function(script, legacyOnly) {
            for(var i=0; i < SCRIPTS.length; i++) if(SCRIPTS[i] === script) return false;
            if((legacyOnly && isLegacy()) || (!legacyOnly)) {
                SCRIPTS.push(script);
            }
        };
    </script>

    <!-- fonts.com - Async Font Loading -->
    <script type="text/javascript">
        (function() {
            var fontsComPath = '//fast.fonts.net/jsapi/0d47d266-cd84-4ef7-adb8-5a44ad7011ef.js',
                    fontsComJS = document.createElement('script');
            fontsComJS.type = 'text/javascript';
            fontsComJS.async = true;
            fontsComJS.src = fontsComPath;
            var head = document.getElementsByTagName("head")[0];
            head.appendChild(fontsComJS);
        })();
    </script><meta name="robots" content="noindex">
</head>
<body data-ajax-domain="192.168.1.13:33351">
</body>
</html>`

var integratedTestHtml = `<html>
  <head>
    <link uic-remove rel="stylesheet" type="text/css" href="testing.css"/>
    <link rel="stylesheet" type="text/css" href="special.css"/>
    <script type="text/uic-meta">
      {
       "foo": "bar",
       "boo": "bazz",
       "categories": ["animal", "human"]
      }
    </script>
    <script uic-remove>
      va xzw = "some test code""
    </script>
    <script src="myScript.js"></script>
  </head>
  <body>
    <ul uic-remove>
      <!-- A Navigation for testing -->
    </ul>
    <uic-fragment name="headline">
      <h1>This is a headline</h1>
    </uic-fragment>
    <uic-fragment name="content">
      Bli Bla blub
      <uic-fetch src="example.com/foo" timeout="100" required="true"/>
      <uic-include src="example.com/foo#content" required="true"/>
      <uic-include src="example.com/optional#content">
        <p>some alternative text</p>
      </uic-include>
      <div uic-remove>
         Some element for testing
      </div>
      <hr/>
      Bli Bla blub
    </uic-fragment>
    <uic-tail>
      <!-- some script tags to insert at the end -->
      <script src="foo.js"></script>
      <script src="bar.js"></script>
      <script uic-remove src="demo.js"></script>
    </uic-tail>
  </body>
</html>
`

var integratedTestHtmlExpectedMeta = map[string]interface{}{
	"foo":        "bar",
	"boo":        "bazz",
	"categories": []interface{}{"animal", "human"},
}

var integratedTestHtmlExpectedHead = `
    <link rel="stylesheet" type="text/css" href="special.css"/>
    <script src="myScript.js"></script>`

var integratedTestHtmlExpectedHeadline = `<h1>This is a headline</h1>`

var integratedTestHtmlExpectedContent = `
      Bli Bla blub
      §[> example.com/foo#content]§
      §[#> example.com/optional#content]§
         <p>some alternative text</p>
      §[/example.com/optional#content]§
      <hr/>
      Bli Bla blub`

var integratedTestHtmlExpectedTail = `
      <!-- some script tags to insert at the end -->
      <script src="foo.js"></script>
      <script src="bar.js"></script>`

func Test_HtmlContentParser_LoadEmptyContent(t *testing.T) {
	a := assert.New(t)

	in := strings.NewReader(`<html>
  <head>
  </head>
  <body>
  </body>
</html>
`)
	c := NewMemoryContent()
	parser := NewHtmlContentParser(true, true)
	err := parser.Parse(c, in)
	a.NoError(err)

	a.Equal(0, len(c.Body()))
	a.Equal(0, len(c.Meta()))
	a.Equal(0, len(c.RequiredContent()))
	a.Nil(c.Head())
	a.Nil(c.Tail())
}

func Test_HtmlContentParser_ParseBrokenScript(t *testing.T) {
	a := assert.New(t)

	in := strings.NewReader(`<html>
  <head>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script> // this script is not closed correctly </ipt>
  </head>
  <body>
  <h1>Crash!</h1>
  </body>
</html>
`)
	c := NewMemoryContent()
	parser := NewHtmlContentParser(true, true)
	err := parser.Parse(c, in)
	a.Error(err)
	a.Contains(err.Error(), "Tag not properly ended. Expected </script>. Error was: EOF")
	a.Equal(0, len(c.Body()))
	a.Equal(0, len(c.Meta()))
	a.Equal(0, len(c.RequiredContent()))
	a.Nil(c.Head())
	a.Nil(c.Tail())
}

func Test_HtmlContentParser_ParseBrokenScript2(t *testing.T) {
	a := assert.New(t)

	in := strings.NewReader(`<html>
  <head>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script>`)
	c := NewMemoryContent()
	parser := NewHtmlContentParser(true, true)
	err := parser.Parse(c, in)
	a.Error(err)
	a.Contains(err.Error(), "expected text node for inline script, but found Error")
	a.Equal(0, len(c.Body()))
	a.Equal(0, len(c.Meta()))
	a.Equal(0, len(c.RequiredContent()))
	a.Nil(c.Head())
	a.Nil(c.Tail())
}

func Test_HtmlContentParser_ParseEmptyScript(t *testing.T) {
	a := assert.New(t)

	in := strings.NewReader(`<html>
  <head>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script></script> <!-- dont treat it as an error -->
  </head>
  <body>
  </body>
</html>
`)
	c := NewMemoryContent()
	parser := NewHtmlContentParser(true, true)
	err := parser.Parse(c, in)
	a.NoError(err)
	a.Equal(0, len(c.Body()))
	a.Equal(0, len(c.Meta()))
	a.Equal(0, len(c.RequiredContent()))
	a.NotNil(c.Head()) // Head contains the (empty) script
	a.Nil(c.Tail())
}

func Test_HtmlContentParser_parseHead_withMultipleMetaTags_and_Titles_and_Canonicals(t *testing.T) {
	a := assert.New(t)

	parser := NewHtmlContentParser(true, true)
	z := html.NewTokenizer(bytes.NewBufferString(productUiGeneratedHtml))

	z.Next()
	c := NewMemoryContent()
	err := parser.parseHead(z, c)
	a.NoError(err)
	containsFragment(t, "<title>navigationservice</title>", c.Head())
}

func Test_HtmlContentParser_parseHead(t *testing.T) {
	a := assert.New(t)

	parser := NewHtmlContentParser(true, true)
	z := html.NewTokenizer(bytes.NewBufferString(`<head>
  <div uic-remove>
    <script>
    sdcsdc
    </script>
  </div>
  <xx/> 
  <foo>xxx</foo>
  <div uic-remove>
    <script>
    sdcsdc
    </script>
  </div> 
  <bar>xxx</bar>
  <script type="text/uic-meta">
      {
       "foo": "bar"
      }
  </script>
  <div uic-remove>
    <script>
    sdcsdc
    </script>
  </div>
</head>
`))

	z.Next() // At <head ..
	c := NewMemoryContent()
	err := parser.parseHead(z, c)
	a.NoError(err)

	eqFragment(t, "<xx/><foo>xxx</foo><bar>xxx</bar>", c.Head())
	a.Equal("bar", c.Meta()["foo"])
}

func Test_HtmlContentParser_collectStylesheets_bodyAsDefaultFragment(t *testing.T) {
	a := assert.New(t)

	parser := NewHtmlContentParser(true, true)
	z := bytes.NewBufferString(`<head>
	<!-- will be found by the HeaderParser -->
	<link rel="stylesheet" href="/navigationservice/stylesheets/main-93174ed18d.css">
	<link rel="stylesheet" href="/navigationservice/stylesheets/main-93174ed18d.css">
	<script src="/rebrush/assets/typo/javascripts/picturefill-f350acdff4.min.js" async></script>
	<script>var test="abc";</script>
	<script src="/rebrush/assets/typo/javascripts/jquery.min.js" async></script>
	<link rel="canonical" href="/baumarkt/bauen-renovieren/suche">
	<link rel="stylesheet" href="/navigationservice/stylesheets/main-03174ed18d.css">
</head>
<body>
	<div>
		<!-- will be found by the BodyParser -->
		<script src="/rebrush/assets/typo/javascripts/picturefill-f350acdff4.min.js" async></script>
		<link rel="stylesheet" href="/productservice/stylesheets/main-93174ed18d.css">
		<script>var test="abc";</script>
		<script src="/rebrush/assets/typo/javascripts/jquery.min.js" async></script>
		<uic-fragment name="content">
			Bli Bla blub
			<script src="/rebrush/assets/typo/javascripts/picturefill-f350acdff4.min.js" async></script>
			<script>var test="abc";</script>
			<link rel="stylesheet" href="/basketservice/stylesheets/main-93174ed18d.css">
			<script src="/rebrush/assets/typo/javascripts/jquery.min.js" async></script>
		</uic-fragment>
	</div>
</body>
`)

	c := NewMemoryContent()
	err := parser.Parse(c, z)
	a.NoError(err)
	// all head link elements were found
	a.Equal("rel=\"stylesheet\" href=\"/navigationservice/stylesheets/main-93174ed18d.css\"",
		joinAttrs(c.Head().LinkTags()[0]))
	a.Equal("rel=\"stylesheet\" href=\"/navigationservice/stylesheets/main-93174ed18d.css\"",
		joinAttrs(c.Head().LinkTags()[1]))
	a.Equal("rel=\"canonical\" href=\"/baumarkt/bauen-renovieren/suche\"",
		joinAttrs(c.Head().LinkTags()[2]))
	a.Equal("rel=\"stylesheet\" href=\"/navigationservice/stylesheets/main-03174ed18d.css\"",
		joinAttrs(c.Head().LinkTags()[3]))
	// all body-fragment link elements were found
	a.Equal("rel=\"stylesheet\" href=\"/basketservice/stylesheets/main-93174ed18d.css\"",
		joinAttrs(c.Body()["content"].LinkTags()[0]))
	a.Equal("rel=\"stylesheet\" href=\"/productservice/stylesheets/main-93174ed18d.css\"",
		joinAttrs(c.Body()[""].LinkTags()[0]))

	// test script tags
	scriptTags := c.Head().ScriptElements()
	a.Equal("src=\"/rebrush/assets/typo/javascripts/picturefill-f350acdff4.min.js\" async=\"\"",
		joinAttrs(scriptTags[0].Attrs))
	a.Equal(`var test="abc";`, string(scriptTags[1].Text))
	a.Equal("src=\"/rebrush/assets/typo/javascripts/jquery.min.js\" async=\"\"",
		joinAttrs(scriptTags[2].Attrs))
	a.Equal(3, len(scriptTags))

	scriptTags = c.Body()[""].ScriptElements()
	a.Equal("src=\"/rebrush/assets/typo/javascripts/picturefill-f350acdff4.min.js\" async=\"\"",
		joinAttrs(scriptTags[0].Attrs))
	a.Equal(`var test="abc";`, string(scriptTags[1].Text))
	a.Equal("src=\"/rebrush/assets/typo/javascripts/jquery.min.js\" async=\"\"",
		joinAttrs(scriptTags[2].Attrs))
	a.Equal(3, len(scriptTags))

	scriptTags = c.Body()["content"].ScriptElements()
	a.Equal("src=\"/rebrush/assets/typo/javascripts/picturefill-f350acdff4.min.js\" async=\"\"",
		joinAttrs(scriptTags[0].Attrs))
	a.Equal(`var test="abc";`, string(scriptTags[1].Text))
	a.Equal("src=\"/rebrush/assets/typo/javascripts/jquery.min.js\" async=\"\"",
		joinAttrs(scriptTags[2].Attrs))
	a.Equal(3, len(scriptTags))

	// assert that scripts were removed from the fragment content
	a.NotContains(c.Body()[""].(*StringFragment).Content(), `<script>var test="abc";</script>`)
	a.NotContains(c.Body()[""].(*StringFragment).Content(), `<script src="/rebrush/assets/typo/javascripts/picturefill-f350acdff4.min.js" async></script>`)
	a.NotContains(c.Body()[""].(*StringFragment).Content(), ` </script>`)
	a.NotContains(c.Head().(*StringFragment).Content(), ` </script>`)
}

func Test_HtmlContentParser_collectStylesheets_OverrideDefault(t *testing.T) {
	a := assert.New(t)

	parser := NewHtmlContentParser(true, true)
	z := bytes.NewBufferString(`<head>
	<!-- will be found by the HeaderParser -->
	<link rel="stylesheet" href="/navigationservice/stylesheets/main-93174ed18d.css">
	<link rel="stylesheet" href="/navigationservice/stylesheets/main-93174ed18d.css">
	<link rel="stylesheet" href="/navigationservice/stylesheets/main-03174ed18d.css">
</head>
<body>
	<div>
		<!-- will be found by the BodyParser -->
		<link rel="stylesheet" href="/productservice/stylesheets/main-93174ed18d.css">
		<uic-fragment name="content">
		Bli Bla blub
		<link rel="stylesheet" href="/basketservice/stylesheets/main-93174ed18d.css">
		</uic-fragment>
		<uic-fragment name="">
		Bli Bla blub
		<link rel="stylesheet" href="/override/stylesheets/main-93174ed18d.css">
		</uic-fragment>
	</div>
</body>
`)

	c := NewMemoryContent()
	err := parser.Parse(c, z)
	a.NoError(err)
	a.Equal(
		"rel=\"stylesheet\" href=\"/basketservice/stylesheets/main-93174ed18d.css\"",
		joinAttrs(c.Body()["content"].LinkTags()[0]))
	a.Equal("rel=\"stylesheet\" href=\"/override/stylesheets/main-93174ed18d.css\"",
		joinAttrs(c.Body()[""].LinkTags()[0]))
}

func Test_HtmlContentParser_parseBody(t *testing.T) {
	a := assert.New(t)

	parser := NewHtmlContentParser(true, true)
	z := html.NewTokenizer(bytes.NewBufferString(`<body some="attribute">
    <h1>Default Fragment Content</h1><br>
    <uic-include src="example.com/xyz" required="true" param-foo="bar" param-bazz="buzz"/>
    <ul uic-remove>
      <!-- A Navigation for testing -->
    </ul>
    <uic-fragment name="headline">
      <h1>Headline</h1>
      <uic-include src="example.com/optional#content"/>
    </uic-fragment>
    <uic-fragment name="content">
      some content
      <uic-include src="example.com/foo#content" required="true" param-bli="bla"/>
      <uic-include src="example.com/optional#content" required="false"/>
      <uic-include src="#local" required="true"/>
    </uic-fragment>
    <uic-tail>
      <!-- tail -->
      <uic-include src="example.com/tail" timeout="100" required="true"/>
    </uic-tail>
  </body>`))

	z.Next() // At <body ..
	c := NewMemoryContent()
	err := parser.parseBody(z, c)
	a.NoError(err)

	a.Equal(3, len(c.Body()))
	eqFragment(t, "<h1>Default Fragment Content</h1><br>\n§[> example.com/xyz]§", c.Body()[""])
	eqFragment(t, `<h1>Headline</h1> §[#> example.com/optional#content]§§[/example.com/optional#content]§`, c.Body()["headline"])
	eqFragment(t, `some content`+
		`§[> example.com/foo#content]§`+
		`§[#> example.com/optional#content]§§[/example.com/optional#content]§`+
		`§[> local]§`, c.Body()["content"])
	eqFragment(t, "<!-- tail -->§[> example.com/tail]§", c.Tail())

	eqFragment(t, `some="attribute"`, c.BodyAttributes())

	a.Equal(5, len(c.Dependencies()))
	a.Equal(c.Dependencies()["example.com/xyz"], Params{"foo": "bar", "bazz": "buzz"})
	a.Equal(c.Dependencies()["example.com/foo"], Params{"bli": "bla"})
	a.Contains(c.Dependencies(), "example.com/foo")
	a.Contains(c.Dependencies(), "example.com/optional")
	a.Contains(c.Dependencies(), "example.com/tail")
	a.Contains(c.Dependencies(), "example.com/xyz")
	a.Contains(c.Dependencies(), "local")
}

func Test_HtmlContentParser_fetchDependencies(t *testing.T) {
	a := assert.New(t)

	parser := NewHtmlContentParser(true, true)
	z := html.NewTokenizer(bytes.NewBufferString(`<body>
           foo
           <uic-fetch src="example.com/foo" timeout="42000" required="true" name="foo"/>
           <uic-fetch src="example.com/optional" timeout="100" required="false"/>
           <uic-fetch src="discovered" discoveredBy="192.168.0.42:8008"/>
         </body>`))

	z.Next() // At <body ..
	c := NewMemoryContent()
	err := parser.parseBody(z, c)
	a.NoError(err)

	eqFragment(t, "foo", c.Body()[""])

	a.Equal(3, len(c.RequiredContent()))
	a.Equal(&FetchDefinition{
		URL:      "example.com/foo",
		Name:     "foo",
		Timeout:  time.Millisecond * 42000,
		Required: true,
	}, c.requiredContent["example.com/foo"])

	a.Equal(&FetchDefinition{
		URL:      "example.com/optional",
		Name:     "example.com/optional",
		Timeout:  time.Millisecond * 100,
		Required: false,
	}, c.requiredContent["example.com/optional"])

	a.True(c.requiredContent["discovered"].ServiceDiscoveryActive)
	a.NotNil(c.requiredContent["discovered"].ServiceDiscovery)

}

func Test_HtmlContentParser_fetchAndInclude_ErrorCases(t *testing.T) {
	a := assert.New(t)

	testCases := []string{
		`<uic-fetch/>`,
		`<uic-fetch src="example.com/foo" required="tr42ue"/>`,
		`<uic-fetch src="example.com/foo" timeout="sdcascdsdc"/>`,
		`<uic-fragment name="bla"><uic-include/><uic-fragment>`,
		`<uic-include src="example.com/foo" required="tr42ue"/>`,
	}

	for i, test := range testCases {
		t.Run(fmt.Sprintf("test #%v", i), func(t *testing.T) {
			parser := NewHtmlContentParser(true, true)
			z := html.NewTokenizer(bytes.NewBufferString(
				"<body>" + test + "</body>",
			))
			z.Next() // At <body ..
			err := parser.parseBody(z, NewMemoryContent())
			a.Error(err)
		})
	}
}

func Test_HtmlContentParser_parseBody_OnlyDefaultFragment(t *testing.T) {
	a := assert.New(t)

	parser := NewHtmlContentParser(true, true)
	z := html.NewTokenizer(bytes.NewBufferString(`<body>
    <h1>Default Fragment Content</h1><br>
    <uic-include src="example.com/foo#content" required="true"/>
  </body>`))

	z.Next() // At <body ..
	c := NewMemoryContent()
	err := parser.parseBody(z, c)
	a.NoError(err)

	a.Equal(1, len(c.Body()))
	eqFragment(t, "<h1>Default Fragment Content</h1><br> §[> example.com/foo#content]§", c.Body()[""])
}

func Test_HtmlContentParser_parseBody_DefaultFragmentOverwritten(t *testing.T) {
	a := assert.New(t)

	parser := NewHtmlContentParser(true, true)
	z := html.NewTokenizer(bytes.NewBufferString(`<body>
    <h1>Default Fragment Content</h1><br>
    <uic-fragment>
      Overwritten
    </uic-fragment>
	<uic-fragment name="content">
      Unused
    </uic-fragment>
  </body>`))

	z.Next() // At <body ..
	c := NewMemoryContent()
	err := parser.parseBody(z, c)
	a.NoError(err)

	a.Equal(2, len(c.Body()))
	eqFragment(t, "Overwritten", c.Body()[""])
	eqFragment(t, "Unused", c.Body()["content"])
}

func Test_HtmlContentParser_parseHead_JsonError(t *testing.T) {
	a := assert.New(t)

	parser := NewHtmlContentParser(true, true)
	z := html.NewTokenizer(bytes.NewBufferString(`
<script type="text/uic-meta">
      {
</script>
`))

	c := NewMemoryContent()
	err := parser.parseHead(z, c)

	a.Error(err)
	a.Contains(err.Error(), "error while parsing json from meta json")
}

func Test_HtmlContentParser_parseFragment(t *testing.T) {
	parser := NewHtmlContentParser(true, true)
	a := assert.New(t)

	z := html.NewTokenizer(bytes.NewBufferString(`<uic-fragment name="content">
      Bli Bla blub
      <link rel="stylesheet" href="/navigationservice/stylesheets/main-93174ed18d.css">
      <Br>
      <uic-include src="example.com/foo#content" required="true"/>
      <uic-include src="example.com/optional#content" required="false"/>
      <div uic-remove>
         <br>
         Some element for testing
      </div>
      <hr/>     
      Bli Bla §[ aVariable ]§ blub
    </uic-fragment><testend>`))

	z.Next() // At <uic-fragment name ..
	f, _, err := parser.parseFragment(z)
	a.NoError(err)

	expected := `Bli Bla blub

      <br>
      §[> example.com/foo#content]§
      §[#> example.com/optional#content]§§[/example.com/optional#content]§
      <hr/>
      Bli Bla §[ aVariable ]§ blub`
	eqFragment(t, expected, f)

	z.Next()
	endTag, _ := z.TagName()
	a.Equal("testend", string(endTag))
	a.Equal(`rel="stylesheet" href="/navigationservice/stylesheets/main-93174ed18d.css"`, joinAttrs(f.LinkTags()[0]))
	a.Equal(1, len(f.LinkTags()))
}

// Regression test: to ensure, that escaped entities in attributes do not lead to a problem.
func Test_HtmlContentParser_parseFragment_EntityAttribute(t *testing.T) {
	a := assert.New(t)

	testHtml := `<a style="text-decoration: none" href="/produktkatalog?&amp;page=91">`
	in := strings.NewReader(`<html><head>` + testHtml + `</head><body>` +
		testHtml + `<uic-fragment name="content">` + testHtml + `</uic-fragment></body></html>`)

	c := NewMemoryContent()
	parser := NewHtmlContentParser(true, true)
	err := parser.Parse(c, in)
	a.NoError(err)

	eqFragment(t, testHtml, c.Head())
	eqFragment(t, testHtml, c.Body()[""])
	eqFragment(t, testHtml, c.Body()["content"])
}

func Test_HtmlContentParser_parseMetaJson(t *testing.T) {
	a := assert.New(t)

	z := html.NewTokenizer(bytes.NewBufferString(`<script type="text/uic-meta">
      {
       "foo": "bar",
       "boo": "bazz",
       "categories": ["animal", "human"]
      }
    </script>`))

	z.Next() // At <script ..
	c := NewMemoryContent()
	err := parseMetaJson(z, c)
	a.NoError(err)

	a.Equal("bar", c.Meta()["foo"])
	a.Equal("bazz", c.Meta()["boo"])
}

func Test_HtmlContentParser_parseMetaJson_Errors(t *testing.T) {
	a := assert.New(t)

	testCases := []struct {
		html      string
		errorText string
	}{
		{
			html:      `<script type="text/uic-meta"></script>`,
			errorText: "expected text node for meta",
		},
		{
			html:      `<script type="text/uic-meta">{"sdc":</script>`,
			errorText: "error while parsing json from meta json",
		},
		{
			html:      `<script type="text/uic-meta">{}`,
			errorText: "Tag not properly ended",
		},
	}

	for _, test := range testCases {
		z := html.NewTokenizer(bytes.NewBufferString(test.html))
		z.Next() // At <script ..
		err := parseMetaJson(z, NewMemoryContent())

		a.Error(err)
		a.Contains(err.Error(), test.errorText)
	}
}

func Test_HtmlContentParser_skipSubtreeIfUicRemove(t *testing.T) {
	a := assert.New(t)

	z := html.NewTokenizer(bytes.NewBufferString(`<a><b uic-remove>
    sdcsdc
    <hr/>
    <br>
    <img src="http://foo">
    <foo>xxx</foo>
    <br/>
</b></a>`))

	z.Next()
	tt := z.Next() // at b
	attrs := readAttributes(z, make([]html.Attribute, 0, 10))
	skipped := skipSubtreeIfUicRemove(z, tt, "b", attrs)

	a.True(skipped)
	token := z.Next()
	a.Equal(html.EndTagToken, token)
	tag, _ := z.TagName()
	a.Equal("a", string(tag))
}

func Test_joinAttrs(t *testing.T) {
	a := assert.New(t)
	a.Equal(``, joinAttrs([]html.Attribute{}))
	a.Equal(`some="attribute"`, joinAttrs([]html.Attribute{{Key: "some", Val: "attribute"}}))
	a.Equal(`a="b" some="attribute"`, joinAttrs([]html.Attribute{{Key: "a", Val: "b"}, {Key: "some", Val: "attribute"}}))
	a.Equal(`a="--&#34;--"`, joinAttrs([]html.Attribute{{Key: "a", Val: `--"--`}}))
	a.Equal(`ns:a="b"`, joinAttrs([]html.Attribute{{Namespace: "ns", Key: "a", Val: "b"}}))
	a.Equal(`async=""`, joinAttrs([]html.Attribute{{Key: "async", Val: ""}}))
}

func eqFragment(t *testing.T, expected string, f Fragment) {
	if f == nil {
		t.Error("Fragment is nil, but expected:", expected)
		return
	}
	sf := f.(*StringFragment).Content()
	sfStripped := strings.Replace(string(sf), " ", "", -1)
	sfStripped = strings.Replace(string(sfStripped), "\n", "", -1)
	expectedStripped := strings.Replace(expected, " ", "", -1)
	expectedStripped = strings.Replace(expectedStripped, "\n", "", -1)

	if expectedStripped != sfStripped {
		t.Error("Fragment is not equal: \nexpected: ", expected, "\nactual: ", sf)
	}
}

func containsFragment(t *testing.T, contained string, f Fragment) {
	if f == nil {
		t.Error("Fragment is nil, but expected:", contained)
		return
	}
	sf := f.(*StringFragment).Content()
	sfStripped := strings.Replace(string(sf), " ", "", -1)
	sfStripped = strings.Replace(string(sfStripped), "\n", "", -1)

	if !strings.Contains(sfStripped, contained) {
		t.Error("Fragment does not contain expected value: \nexpected: ", contained, "\nactual:  ", sf)
	}
}

func Test_ParseHeadFragment_Filter_Title(t *testing.T) {
	a := assert.New(t)

	originalHeadString := `<meta charset="utf-8">
	<title>navigationservice</title>



	<!-- START Include legacy styles - emulate integration -->

	<!-- END Include legacy styles -->

	<!-- START Include jquery lib - add to SCRIPTS again after last JS from legacy system is removed -->

	<!-- END Include jquery lib -->

	<link rel="stylesheet" href="/navigationservice/stylesheets/main-93174ed18d.css">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">

	<script>
	// Define global SCRIPTS variable and
	// global loadScript() method to loading scripts
	// async but in order.
	// Each module register it's javascript by calling
	// this method:
	//
	// loadScript('/navigationservice/components/molecules/teaser/teaser.js');
	//
		SCRIPTS = ['/navigationservice/javascripts/main-e566a7bb73.js'];
	isLegacy = function() {
		return typeof Object.assign === 'function' ? false : true;
	};

	loadScript = function(script, legacyOnly) {
		for(var i=0; i < SCRIPTS.length; i++) if(SCRIPTS[i] === script) return false;
		if((legacyOnly && isLegacy()) || (!legacyOnly)) {
		SCRIPTS.push(script);
		}
	};
	</script>

	<!-- fonts.com - Async Font Loading -->`

	expectedParsedHead := `<meta charset="utf-8">




	<!-- START Include legacy styles - emulate integration -->

	<!-- END Include legacy styles -->

	<!-- START Include jquery lib - add to SCRIPTS again after last JS from legacy system is removed -->

	<!-- END Include jquery lib -->

	<link rel="stylesheet" href="/navigationservice/stylesheets/main-93174ed18d.css">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">

	<script>
	// Define global SCRIPTS variable and
	// global loadScript() method to loading scripts
	// async but in order.
	// Each module register it's javascript by calling
	// this method:
	//
	// loadScript('/navigationservice/components/molecules/teaser/teaser.js');
	//
		SCRIPTS = ['/navigationservice/javascripts/main-e566a7bb73.js'];
	isLegacy = function() {
		return typeof Object.assign === 'function' ? false : true;
	};

	loadScript = function(script, legacyOnly) {
		for(var i=0; i < SCRIPTS.length; i++) if(SCRIPTS[i] === script) return false;
		if((legacyOnly && isLegacy()) || (!legacyOnly)) {
		SCRIPTS.push(script);
		}
	};
	</script>

	<!-- fonts.com - Async Font Loading -->`

	headPropertyMap := make(map[string]string)
	headPropertyMap["title"] = "title"
	headFragment := NewStringFragment(originalHeadString)

	ParseHeadFragment(headFragment, headPropertyMap)

	expectedParsedHead = removeTabsAndNewLines(expectedParsedHead)
	resultString := removeTabsAndNewLines(headFragment.Content())

	a.Equal(expectedParsedHead, resultString)
}

func Test_ParseHeadFragment_Filter_Meta_Tag(t *testing.T) {
	a := assert.New(t)

	originalHeadString := `<meta charset="utf-8">

	<title>navigationservice</title>



	<!-- START Include legacy styles - emulate integration -->

	<!-- END Include legacy styles -->

	<!-- START Include jquery lib - add to SCRIPTS again after last JS from legacy system is removed -->

	<!-- END Include jquery lib -->

	<link rel="stylesheet" href="/navigationservice/stylesheets/main-93174ed18d.css">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta name="blub" content="width=device-width, initial-scale=1.0">
	<script>
	// Define global SCRIPTS variable and
	// global loadScript() method to loading scripts
	// async but in order.
	// Each module register it's javascript by calling
	// this method:
	//
	// loadScript('/navigationservice/components/molecules/teaser/teaser.js');
	//
		SCRIPTS = ['/navigationservice/javascripts/main-e566a7bb73.js'];
	isLegacy = function() {
		return typeof Object.assign === 'function' ? false : true;
	};

	loadScript = function(script, legacyOnly) {
		for(var i=0; i < SCRIPTS.length; i++) if(SCRIPTS[i] === script) return false;
		if((legacyOnly && isLegacy()) || (!legacyOnly)) {
		SCRIPTS.push(script);
		}
	};
	</script>

	<!-- fonts.com - Async Font Loading -->`

	expectedParsedHead := `
	<title>navigationservice</title>



	<!-- START Include legacy styles - emulate integration -->

	<!-- END Include legacy styles -->

	<!-- START Include jquery lib - add to SCRIPTS again after last JS from legacy system is removed -->

	<!-- END Include jquery lib -->

	<link rel="stylesheet" href="/navigationservice/stylesheets/main-93174ed18d.css">
	<meta name="blub" content="width=device-width, initial-scale=1.0">

	<script>
	// Define global SCRIPTS variable and
	// global loadScript() method to loading scripts
	// async but in order.
	// Each module register it's javascript by calling
	// this method:
	//
	// loadScript('/navigationservice/components/molecules/teaser/teaser.js');
	//
		SCRIPTS = ['/navigationservice/javascripts/main-e566a7bb73.js'];
	isLegacy = function() {
		return typeof Object.assign === 'function' ? false : true;
	};

	loadScript = function(script, legacyOnly) {
		for(var i=0; i < SCRIPTS.length; i++) if(SCRIPTS[i] === script) return false;
		if((legacyOnly && isLegacy()) || (!legacyOnly)) {
		SCRIPTS.push(script);
		}
	};
	</script>

	<!-- fonts.com - Async Font Loading -->`

	headMetaPropertyMap := make(map[string]string)
	headMetaPropertyMap["meta_charset"] = "whatever"
	headMetaPropertyMap["meta_name_viewport"] = "already_exists"

	headFragment := NewStringFragment(originalHeadString)
	ParseHeadFragment(headFragment, headMetaPropertyMap)

	expectedParsedHead = removeTabsAndNewLines(expectedParsedHead)
	resultString := removeTabsAndNewLines(headFragment.Content())

	a.Equal(expectedParsedHead, resultString)
}

func Test_ParseHeadFragment_Filter_Link_Canonical_Tag(t *testing.T) {
	a := assert.New(t)

	// GIVEN
	originalHeadString := `<meta charset="utf-8">

	<link rel="canonical" href="/navigationservice">

	<title>navigationservice</title>
	<!-- START Include jquery lib - add to SCRIPTS again after last JS from legacy system is removed -->
	<!-- END Include jquery lib -->
	<link rel="stylesheet" href="/navigationservice/stylesheets/main-93174ed18d.css">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta name="blub" content="width=device-width, initial-scale=1.0">
	<script>
	// Define global SCRIPTS variable and
	// global loadScript() method to loading scripts
	// async but in order.
	// Each module register it's javascript by calling
	// this method:
	//
	// loadScript('/navigationservice/components/molecules/teaser/teaser.js');
	//
		SCRIPTS = ['/navigationservice/javascripts/main-e566a7bb73.js'];
	isLegacy = function() {
		return typeof Object.assign === 'function' ? false : true;
	};
	loadScript = function(script, legacyOnly) {
		for(var i=0; i < SCRIPTS.length; i++) if(SCRIPTS[i] === script) return false;
		if((legacyOnly && isLegacy()) || (!legacyOnly)) {
		SCRIPTS.push(script);
		}
	};
	</script>
	<!-- fonts.com - Async Font Loading -->`

	expectedParsedHead := `<meta charset="utf-8">



	<title>navigationservice</title>
	<!-- START Include jquery lib - add to SCRIPTS again after last JS from legacy system is removed -->
	<!-- END Include jquery lib -->
	<link rel="stylesheet" href="/navigationservice/stylesheets/main-93174ed18d.css">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta name="blub" content="width=device-width, initial-scale=1.0">
	<script>
	// Define global SCRIPTS variable and
	// global loadScript() method to loading scripts
	// async but in order.
	// Each module register it's javascript by calling
	// this method:
	//
	// loadScript('/navigationservice/components/molecules/teaser/teaser.js');
	//
		SCRIPTS = ['/navigationservice/javascripts/main-e566a7bb73.js'];
	isLegacy = function() {
		return typeof Object.assign === 'function' ? false : true;
	};
	loadScript = function(script, legacyOnly) {
		for(var i=0; i < SCRIPTS.length; i++) if(SCRIPTS[i] === script) return false;
		if((legacyOnly && isLegacy()) || (!legacyOnly)) {
		SCRIPTS.push(script);
		}
	};
	</script>
	<!-- fonts.com - Async Font Loading -->`

	headMetaPropertyMap := make(map[string]string)
	headMetaPropertyMap["canonical"] = "/baumarkt/suche"

	headFragment := NewStringFragment(originalHeadString)
	// WHEN
	ParseHeadFragment(headFragment, headMetaPropertyMap)

	// THEN
	expectedParsedHead = removeTabsAndNewLines(expectedParsedHead)
	resultString := removeTabsAndNewLines(headFragment.Content())

	a.Equal(expectedParsedHead, resultString)
}

func Test_ParseHeadFragment_Filter_Link_Canonical_Tag_without_existing_Map(t *testing.T) {
	// GIVEN
	a := assert.New(t)

	originalHeadString := `
	<link rel="stylesheet" href="/searchservice/stylesheets/main-36b9f2e88a.css">
	<link />
	<link rel="canonical" href="/baumarkt/bauen-renovieren/suche">
        <meta charset="utf-8" />
	<link rel="canonical"
	        href="/navigationservice">
	<foo bar=""/>
	<link    href="/feedbackservice"   rel="canonical">
	<title>navigationservice</title>
	`

	expectedParsedHead := `
	<link rel="stylesheet" href="/searchservice/stylesheets/main-36b9f2e88a.css">
	<link />
	<link rel="canonical" href="/baumarkt/bauen-renovieren/suche">
        <meta charset="utf-8" />
	<foo bar=""/>
	<title>navigationservice</title>
	`

	headMetaPropertyMap := make(map[string]string)

	headFragment := NewStringFragment(originalHeadString)
	// WHEN
	ParseHeadFragment(headFragment, headMetaPropertyMap)

	// THEN
	expectedParsedHead = removeTabsAndNewLines(expectedParsedHead)
	resultString := removeTabsAndNewLines(headFragment.Content())

	a.Equal(expectedParsedHead, resultString)
}

func Test_getTag(t *testing.T) {
	a := assert.New(t)

	vals := []struct {
		tagType TagType
		string
	}{
		{LINK, `<link rel="apple-touch-icon" sizes="76x76" href="/productservice/favicons/apple-touch-icon-76x76-1499095290-5d8490ac47.png">`},
		{LINK, `<LiNk rel="apple-touch-icon" sizes="76x76" href="/productservice/favicons/apple-touch-icon-76x76-1499095290-5d8490ac47.png">`},
		{META, `<meta charset="utf-8">`},
		{SCRIPT, `<script type="text/javascript" src="/productservice/javascripts/loadscript-1499095290-572cdaf4d3.js"></script>`},
		{SCRIPT, `<Script type="text/javascript" src="/productservice/javascripts/loadscript-1499095290-572cdaf4d3.js"></script>`},
		{SCRIPT_INLINE, `<script type="text/javascript">function inlineJS(){console.log('hello');}</script>`},
	}

	for _, v := range vals {
		z := html.NewTokenizer(bytes.NewBufferString(v.string))
		_ = z.Next()
		tag, _ := z.TagName()
		attrs := readAttributes(z, make([]html.Attribute, 0, 10))
		_, typ := getTag(tag, attrs)
		a.Equal(v.tagType, typ)
	}
}

func removeTabsAndNewLines(stringToProcess string) string {
	stringToProcess = strings.Replace(stringToProcess, "\n", "", -1)
	stringToProcess = strings.Replace(stringToProcess, "\t", "", -1)
	return stringToProcess
}
