# Weph (Web Cipher)

Weph is a novel, experimental cipher. It's based on classic book ciphers and uses web page instead of a book to encode the message. This is a proof of concept implementation. 

## Theory 

The cipher considers the structure as well as the content of the HTML page when creating the references. The resulting reference objects are used to encode the message. Multiple HTML pages may be used in the generation of the encoded message. When at all possible, the cipher will attempt to act homophonically and use different reference objects for the same letters. 

### Page Reference Object

A page reference object is an index of a textual node found within the HTML page, excluding script and style tag content. 

````
ReferenceObject
{
	// The zero-based index of the what url was used to generate this reference
	unsigned WORD url

	// The tag depth of the content
	unsigned WORD depth

	// A zero-based index of the reference itself, local to the url index.
	unsigned WORD reference_index

	// Textual content of the text node from the HTML
	string text
}
````

## License

[Unlicense](http://unlicense.org/UNLICENSE). This is a Public Domain work. 

[![Public Domain](https://licensebuttons.net/p/mark/1.0/88x31.png)](http://questioncopyright.org/promise)

> ["Make art not law"](http://questioncopyright.org/make_art_not_law_interview) -Nina Paley
