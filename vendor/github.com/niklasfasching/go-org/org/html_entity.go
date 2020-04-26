package org

import (
	"code.gitea.io/gitea/traceinit"
	"strings"
)

var htmlEntityReplacer *strings.Replacer

func init () {
traceinit.Trace("vendor/github.com/niklasfasching/go-org/org/html_entity.go")




	htmlEntities = append(htmlEntities,
		"---", "—",
		"--", "–",
		"...", "…",
	)
	htmlEntityReplacer = strings.NewReplacer(htmlEntities...)
}

/*
Generated & copied over using the following elisp
(Setting up go generate seems like a waste for now - I call YAGNI on that one)

(insert (mapconcat
         (lambda (entity) (concat "`\\" (car entity) "`, `" (nth 6 entity) "`")) ; entity -> utf8
         (remove-if-not 'listp org-entities)
         ",\n"))
*/
var htmlEntities = []string{
	`\Agrave`, `À`,
	`\agrave`, `à`,
	`\Aacute`, `Á`,
	`\aacute`, `á`,
	`\Acirc`, `Â`,
	`\acirc`, `â`,
	`\Amacr`, `Ã`,
	`\amacr`, `ã`,
	`\Atilde`, `Ã`,
	`\atilde`, `ã`,
	`\Auml`, `Ä`,
	`\auml`, `ä`,
	`\Aring`, `Å`,
	`\AA`, `Å`,
	`\aring`, `å`,
	`\AElig`, `Æ`,
	`\aelig`, `æ`,
	`\Ccedil`, `Ç`,
	`\ccedil`, `ç`,
	`\Egrave`, `È`,
	`\egrave`, `è`,
	`\Eacute`, `É`,
	`\eacute`, `é`,
	`\Ecirc`, `Ê`,
	`\ecirc`, `ê`,
	`\Euml`, `Ë`,
	`\euml`, `ë`,
	`\Igrave`, `Ì`,
	`\igrave`, `ì`,
	`\Iacute`, `Í`,
	`\iacute`, `í`,
	`\Icirc`, `Î`,
	`\icirc`, `î`,
	`\Iuml`, `Ï`,
	`\iuml`, `ï`,
	`\Ntilde`, `Ñ`,
	`\ntilde`, `ñ`,
	`\Ograve`, `Ò`,
	`\ograve`, `ò`,
	`\Oacute`, `Ó`,
	`\oacute`, `ó`,
	`\Ocirc`, `Ô`,
	`\ocirc`, `ô`,
	`\Otilde`, `Õ`,
	`\otilde`, `õ`,
	`\Ouml`, `Ö`,
	`\ouml`, `ö`,
	`\Oslash`, `Ø`,
	`\oslash`, `ø`,
	`\OElig`, `Œ`,
	`\oelig`, `œ`,
	`\Scaron`, `Š`,
	`\scaron`, `š`,
	`\szlig`, `ß`,
	`\Ugrave`, `Ù`,
	`\ugrave`, `ù`,
	`\Uacute`, `Ú`,
	`\uacute`, `ú`,
	`\Ucirc`, `Û`,
	`\ucirc`, `û`,
	`\Uuml`, `Ü`,
	`\uuml`, `ü`,
	`\Yacute`, `Ý`,
	`\yacute`, `ý`,
	`\Yuml`, `Ÿ`,
	`\yuml`, `ÿ`,
	`\fnof`, `ƒ`,
	`\real`, `ℜ`,
	`\image`, `ℑ`,
	`\weierp`, `℘`,
	`\ell`, `ℓ`,
	`\imath`, `ı`,
	`\jmath`, `ȷ`,
	`\Alpha`, `Α`,
	`\alpha`, `α`,
	`\Beta`, `Β`,
	`\beta`, `β`,
	`\Gamma`, `Γ`,
	`\gamma`, `γ`,
	`\Delta`, `Δ`,
	`\delta`, `δ`,
	`\Epsilon`, `Ε`,
	`\epsilon`, `ε`,
	`\varepsilon`, `ε`,
	`\Zeta`, `Ζ`,
	`\zeta`, `ζ`,
	`\Eta`, `Η`,
	`\eta`, `η`,
	`\Theta`, `Θ`,
	`\theta`, `θ`,
	`\thetasym`, `ϑ`,
	`\vartheta`, `ϑ`,
	`\Iota`, `Ι`,
	`\iota`, `ι`,
	`\Kappa`, `Κ`,
	`\kappa`, `κ`,
	`\Lambda`, `Λ`,
	`\lambda`, `λ`,
	`\Mu`, `Μ`,
	`\mu`, `μ`,
	`\nu`, `ν`,
	`\Nu`, `Ν`,
	`\Xi`, `Ξ`,
	`\xi`, `ξ`,
	`\Omicron`, `Ο`,
	`\omicron`, `ο`,
	`\Pi`, `Π`,
	`\pi`, `π`,
	`\Rho`, `Ρ`,
	`\rho`, `ρ`,
	`\Sigma`, `Σ`,
	`\sigma`, `σ`,
	`\sigmaf`, `ς`,
	`\varsigma`, `ς`,
	`\Tau`, `Τ`,
	`\Upsilon`, `Υ`,
	`\upsih`, `ϒ`,
	`\upsilon`, `υ`,
	`\Phi`, `Φ`,
	`\phi`, `ɸ`,
	`\varphi`, `φ`,
	`\Chi`, `Χ`,
	`\chi`, `χ`,
	`\acutex`, `𝑥́`,
	`\Psi`, `Ψ`,
	`\psi`, `ψ`,
	`\tau`, `τ`,
	`\Omega`, `Ω`,
	`\omega`, `ω`,
	`\piv`, `ϖ`,
	`\varpi`, `ϖ`,
	`\partial`, `∂`,
	`\alefsym`, `ℵ`,
	`\aleph`, `ℵ`,
	`\gimel`, `ℷ`,
	`\beth`, `ב`,
	`\dalet`, `ד`,
	`\ETH`, `Ð`,
	`\eth`, `ð`,
	`\THORN`, `Þ`,
	`\thorn`, `þ`,
	`\dots`, `…`,
	`\cdots`, `⋯`,
	`\hellip`, `…`,
	`\middot`, `·`,
	`\iexcl`, `¡`,
	`\iquest`, `¿`,
	`\shy`, ``,
	`\ndash`, `–`,
	`\mdash`, `—`,
	`\quot`, `"`,
	`\acute`, `´`,
	`\ldquo`, `“`,
	`\rdquo`, `”`,
	`\bdquo`, `„`,
	`\lsquo`, `‘`,
	`\rsquo`, `’`,
	`\sbquo`, `‚`,
	`\laquo`, `«`,
	`\raquo`, `»`,
	`\lsaquo`, `‹`,
	`\rsaquo`, `›`,
	`\circ`, `∘`,
	`\vert`, `|`,
	`\vbar`, `|`,
	`\brvbar`, `¦`,
	`\S`, `§`,
	`\sect`, `§`,
	`\amp`, `&`,
	`\lt`, `<`,
	`\gt`, `>`,
	`\tilde`, `~`,
	`\slash`, `/`,
	`\plus`, `+`,
	`\under`, `_`,
	`\equal`, `=`,
	`\asciicirc`, `^`,
	`\dagger`, `†`,
	`\dag`, `†`,
	`\Dagger`, `‡`,
	`\ddag`, `‡`,
	`\nbsp`, ` `,
	`\ensp`, ` `,
	`\emsp`, ` `,
	`\thinsp`, ` `,
	`\curren`, `¤`,
	`\cent`, `¢`,
	`\pound`, `£`,
	`\yen`, `¥`,
	`\euro`, `€`,
	`\EUR`, `€`,
	`\dollar`, `$`,
	`\USD`, `$`,
	`\copy`, `©`,
	`\reg`, `®`,
	`\trade`, `™`,
	`\minus`, `−`,
	`\pm`, `±`,
	`\plusmn`, `±`,
	`\times`, `×`,
	`\frasl`, `⁄`,
	`\colon`, `:`,
	`\div`, `÷`,
	`\frac12`, `½`,
	`\frac14`, `¼`,
	`\frac34`, `¾`,
	`\permil`, `‰`,
	`\sup1`, `¹`,
	`\sup2`, `²`,
	`\sup3`, `³`,
	`\radic`, `√`,
	`\sum`, `∑`,
	`\prod`, `∏`,
	`\micro`, `µ`,
	`\macr`, `¯`,
	`\deg`, `°`,
	`\prime`, `′`,
	`\Prime`, `″`,
	`\infin`, `∞`,
	`\infty`, `∞`,
	`\prop`, `∝`,
	`\propto`, `∝`,
	`\not`, `¬`,
	`\neg`, `¬`,
	`\land`, `∧`,
	`\wedge`, `∧`,
	`\lor`, `∨`,
	`\vee`, `∨`,
	`\cap`, `∩`,
	`\cup`, `∪`,
	`\smile`, `⌣`,
	`\frown`, `⌢`,
	`\int`, `∫`,
	`\therefore`, `∴`,
	`\there4`, `∴`,
	`\because`, `∵`,
	`\sim`, `∼`,
	`\cong`, `≅`,
	`\simeq`, `≅`,
	`\asymp`, `≈`,
	`\approx`, `≈`,
	`\ne`, `≠`,
	`\neq`, `≠`,
	`\equiv`, `≡`,
	`\triangleq`, `≜`,
	`\le`, `≤`,
	`\leq`, `≤`,
	`\ge`, `≥`,
	`\geq`, `≥`,
	`\lessgtr`, `≶`,
	`\lesseqgtr`, `⋚`,
	`\ll`, `≪`,
	`\Ll`, `⋘`,
	`\lll`, `⋘`,
	`\gg`, `≫`,
	`\Gg`, `⋙`,
	`\ggg`, `⋙`,
	`\prec`, `≺`,
	`\preceq`, `≼`,
	`\preccurlyeq`, `≼`,
	`\succ`, `≻`,
	`\succeq`, `≽`,
	`\succcurlyeq`, `≽`,
	`\sub`, `⊂`,
	`\subset`, `⊂`,
	`\sup`, `⊃`,
	`\supset`, `⊃`,
	`\nsub`, `⊄`,
	`\sube`, `⊆`,
	`\nsup`, `⊅`,
	`\supe`, `⊇`,
	`\setminus`, `⧵`,
	`\forall`, `∀`,
	`\exist`, `∃`,
	`\exists`, `∃`,
	`\nexist`, `∄`,
	`\nexists`, `∄`,
	`\empty`, `∅`,
	`\emptyset`, `∅`,
	`\isin`, `∈`,
	`\in`, `∈`,
	`\notin`, `∉`,
	`\ni`, `∋`,
	`\nabla`, `∇`,
	`\ang`, `∠`,
	`\angle`, `∠`,
	`\perp`, `⊥`,
	`\parallel`, `∥`,
	`\sdot`, `⋅`,
	`\cdot`, `⋅`,
	`\lceil`, `⌈`,
	`\rceil`, `⌉`,
	`\lfloor`, `⌊`,
	`\rfloor`, `⌋`,
	`\lang`, `⟨`,
	`\rang`, `⟩`,
	`\langle`, `⟨`,
	`\rangle`, `⟩`,
	`\hbar`, `ℏ`,
	`\mho`, `℧`,
	`\larr`, `←`,
	`\leftarrow`, `←`,
	`\gets`, `←`,
	`\lArr`, `⇐`,
	`\Leftarrow`, `⇐`,
	`\uarr`, `↑`,
	`\uparrow`, `↑`,
	`\uArr`, `⇑`,
	`\Uparrow`, `⇑`,
	`\rarr`, `→`,
	`\to`, `→`,
	`\rightarrow`, `→`,
	`\rArr`, `⇒`,
	`\Rightarrow`, `⇒`,
	`\darr`, `↓`,
	`\downarrow`, `↓`,
	`\dArr`, `⇓`,
	`\Downarrow`, `⇓`,
	`\harr`, `↔`,
	`\leftrightarrow`, `↔`,
	`\hArr`, `⇔`,
	`\Leftrightarrow`, `⇔`,
	`\crarr`, `↵`,
	`\hookleftarrow`, `↵`,
	`\arccos`, `arccos`,
	`\arcsin`, `arcsin`,
	`\arctan`, `arctan`,
	`\arg`, `arg`,
	`\cos`, `cos`,
	`\cosh`, `cosh`,
	`\cot`, `cot`,
	`\coth`, `coth`,
	`\csc`, `csc`,
	`\deg`, `deg`,
	`\det`, `det`,
	`\dim`, `dim`,
	`\exp`, `exp`,
	`\gcd`, `gcd`,
	`\hom`, `hom`,
	`\inf`, `inf`,
	`\ker`, `ker`,
	`\lg`, `lg`,
	`\lim`, `lim`,
	`\liminf`, `liminf`,
	`\limsup`, `limsup`,
	`\ln`, `ln`,
	`\log`, `log`,
	`\max`, `max`,
	`\min`, `min`,
	`\Pr`, `Pr`,
	`\sec`, `sec`,
	`\sin`, `sin`,
	`\sinh`, `sinh`,
	`\sup`, `sup`,
	`\tan`, `tan`,
	`\tanh`, `tanh`,
	`\bull`, `•`,
	`\bullet`, `•`,
	`\star`, `⋆`,
	`\lowast`, `∗`,
	`\ast`, `*`,
	`\odot`, `ʘ`,
	`\oplus`, `⊕`,
	`\otimes`, `⊗`,
	`\check`, `✓`,
	`\checkmark`, `✓`,
	`\para`, `¶`,
	`\ordf`, `ª`,
	`\ordm`, `º`,
	`\cedil`, `¸`,
	`\oline`, `‾`,
	`\uml`, `¨`,
	`\zwnj`, `‌`,
	`\zwj`, `‍`,
	`\lrm`, `‎`,
	`\rlm`, `‏`,
	`\smiley`, `☺`,
	`\blacksmile`, `☻`,
	`\sad`, `☹`,
	`\frowny`, `☹`,
	`\clubs`, `♣`,
	`\clubsuit`, `♣`,
	`\spades`, `♠`,
	`\spadesuit`, `♠`,
	`\hearts`, `♥`,
	`\heartsuit`, `♥`,
	`\diams`, `◆`,
	`\diamondsuit`, `◆`,
	`\diamond`, `◆`,
	`\Diamond`, `◆`,
	`\loz`, `⧫`,
	`\_ `, ` `,
	`\_  `, `  `,
	`\_   `, `   `,
	`\_    `, `    `,
	`\_     `, `     `,
	`\_      `, `      `,
	`\_       `, `       `,
	`\_        `, `        `,
	`\_         `, `         `,
	`\_          `, `          `,
	`\_           `, `           `,
	`\_            `, `            `,
	`\_             `, `             `,
	`\_              `, `              `,
	`\_               `, `               `,
	`\_                `, `                `,
	`\_                 `, `                 `,
	`\_                  `, `                  `,
	`\_                   `, `                   `,
	`\_                    `, `                    `,
}
