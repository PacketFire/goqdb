package controllers

var (
	// frontend default page size
	VIEW_SIZE_DEFAULT = 5

	// limit the amount of entries that can be returned,
	// this is also a page size 
	VIEW_SIZE_MAX = 4096

	// quote max char limit
	INPUT_QUOTE_MAX  = 1024

	// tag list items max
	INPUT_TAG_LIST_MAX = 32

	// tag entry character max
	INPUT_TAG_MAX = 32

	// max chars for the input field
	//INPUT_TAG_CHAR_MAX = INPUT_TAG_LIST_MAX * INPUT_TAG_ENTRY_CHAR_MAX

	// max author chars
	INPUT_AUTHOR_MAX = 256

	// form input tag delimiter
	INPUT_TAG_DELIM = " "

	// max number of entries displayed in the tag cloud
	TAG_CLOUD_MAX = 10
)

