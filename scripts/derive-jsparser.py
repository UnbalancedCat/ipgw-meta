#!/usr/bin/env python3
"""Reproduce/check the bounded derivative from an authenticated module cache."""
import argparse, hashlib, json
from pathlib import Path

VERSION='v0.0.0-20261008201629-63c541af38f8'
def derive(name,text):
    if name=='parser.go':
        text=text.replace('type _parser struct {','type _parser struct {\n guard *Guard\n comments []Span',1)
        text=text.replace('func (self *_parser) enterNesting() {','func (self *_parser) enterNesting() {\n if self.guard != nil && self.depth >= 64 { panic(stop{"depth_limit"}) }',1)
        # Source maps stay disabled even if a future internal caller omits an option.
        text=text.replace('\treturn p\n}', '\tp.opts.disableSourceMaps = true\n\treturn p\n}',1)
    if name=='lexer.go':
        text=text.replace('func (self *_parser) read() {','func (self *_parser) read() {\n if self.guard != nil { self.guard.read() }',1)
        sig='func (self *_parser) scan() (tkn token.Token, literal string, parsedLiteral unistring.String, idx file.Idx) {'
        assert text.count(sig)==1
        text=text.replace(sig,sig+'\n if self.guard != nil { self.guard.token() }',1)
        for sig in ['func (self *_parser) skipSingleLineComment() {','func (self *_parser) skipMultiLineComment() (hasLineTerminator bool) {']:
            assert text.count(sig)==1
            text=text.replace(sig,sig+'\n start := self.chrOffset-1\n if self.guard != nil { self.guard.token(); defer func(){ self.comments=append(self.comments,Span{start,self.chrOffset,"comment"}) }() }',1)
        sig='func (self *_parser) scanIdentifier() (string, unistring.String, bool, string) {'
        text=text.replace(sig,sig+'\n start := self.chrOffset\n if self.guard != nil { defer func(){ self.comments=append(self.comments,Span{start,self.chrOffset,"identifier"}) }() }',1)
        marker='func (self *_parser) parseTemplateCharacters()'
        at=text.index('\tvar end int',text.index(marker))+len('\tvar end int')
        text=text[:at]+'\n if self.guard != nil { defer func(){ if err=="" { self.comments=append(self.comments,Span{offset,end,"template"}) } }() }'+text[at:]
        # Speculative lookahead must not publish discarded comment intervals.
        text=text.replace('type parserState struct {','type parserState struct {\n commentMark int',1)
        sig='func (self *_parser) mark(state *parserState) *parserState {'
        text=text.replace(sig,sig+'\n',1)
        # mark may allocate a nil state; attach position after that allocation.
        marker='state.idx, state.prevTokenEnd'
        at=text.index(marker,text.index(sig))
        text=text[:at]+'state.commentMark = len(self.comments)\n\t'+text[at:]
        sig='func (self *_parser) restore(state *parserState) {'
        text=text.replace(sig,sig+'\n if state.commentMark <= len(self.comments) { self.comments=self.comments[:state.commentMark] }',1)
        sig='func (self *_parser) peek() token.Token {'
        text=text.replace(sig,sig+'\n mark:=len(self.comments)\n defer func(){self.comments=self.comments[:mark]}()',1)
    if name=='expression.go':
        start=text.index('func (self *_parser) parseRegExpLiteral()')
        end=text.index('\nfunc ',start+1)
        part=text[start:end]
        part=part.replace('endOffset := self.chrOffset','endOffset := self.chrOffset\n exactEnd := self.chrOffset\n if self.guard != nil && err!="" {panic(stop{"syntax_rejected"})}\n if self.guard != nil {defer func(){self.comments=append(self.comments,Span{offset,exactEnd,"regexp"})}()}',1)
        part=part.replace('flags = self.literal','flags = self.literal\n exactEnd = self.chrOffset',1)
        text=text[:start]+part+text[end:]
    return text

def main():
    p=argparse.ArgumentParser();p.add_argument('--upstream',required=True);p.add_argument('--write',action='store_true');args=p.parse_args()
    source=Path(args.upstream);dest=Path(__file__).resolve().parents[1]/'internal/jsparser'
    assert source.name=='goja@'+VERSION
    records=[]
    for item in sorted((source/'parser').glob('*.go')):
        if item.name.endswith('_test.go'):continue
        original=item.read_bytes();out=derive(item.name,original.decode()).encode()
        target=dest/item.name
        if args.write:target.write_bytes(out)
        else:assert target.read_bytes()==out,'derived source drift'
        records.append(dict(file=item.name,upstream_sha256=hashlib.sha256(original).hexdigest(),derived_sha256=hashlib.sha256(out).hexdigest()))
    data=(json.dumps(dict(module='github.com/dop251/goja',version=VERSION,module_sum='h1:OrBT08c059KSn2HsTX43Xb1hx3b81UxNmJ9mprCySIs=',files=records),indent=2)+'\n').encode()
    for name,value in [('upstream.json',data),('LICENSE', (source/'LICENSE').read_bytes())]:
        if args.write:(dest/name).write_bytes(value)
        else:assert (dest/name).read_bytes()==value
    print(json.dumps(dict(files=len(records),matches=True)))
if __name__=='__main__':main()
