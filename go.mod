module webtyp.com/layout

go 1.26.8

require (
	webtyp.com/components v0.8.16
	webtyp.com/css v0.4.29
	webtyp.com/date v0.0.9
	webtyp.com/dom v0.13.22
	webtyp.com/fmt v1.0.0
	webtyp.com/form v0.4.29
	webtyp.com/html v0.0.24
	webtyp.com/image v0.1.16
	webtyp.com/input v0.0.18
	webtyp.com/lang v0.1.3
	webtyp.com/model v0.2.2
	webtyp.com/msgtype v0.1.0
	webtyp.com/svg v0.3.14
	webtyp.com/time v0.5.7
	webtyp.com/view v0.6.27
	webtyp.com/widget v0.6.37
)

require webtyp.com/escape v0.1.0 // indirect

// TEMPORAL — solo para probar en el iPhone el fix de Focus(preventScroll) en
// dom. Quitar y volver a la versión publicada en cuanto el test manual confirme
// el resultado (funcione o no).

require (
	webtyp.com/color v0.1.2 // indirect
	webtyp.com/context v0.0.23 // indirect
	webtyp.com/font v0.0.5 // indirect
	webtyp.com/icons v0.0.7
	webtyp.com/js v0.1.1 // indirect
	webtyp.com/json v0.5.29 // indirect
	webtyp.com/router v0.3.2 // indirect
)
