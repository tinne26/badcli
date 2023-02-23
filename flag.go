package badcli

type flag struct {
	//Name string // to be used with --
	Value FlagValue
	Usage string
	SetByUser bool
}
