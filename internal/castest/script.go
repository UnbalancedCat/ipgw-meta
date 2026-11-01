// Package castest constructs synthetic-only protocol fixtures. Never import it from product code.
package castest

import "strconv"

func LoginScript(key string) string {
	return `function syntheticLogin(){var $u=$("#un"),$p=$("#pd");var u=$u.val();var p=$p.val();if(!u){return;}if(!p){return;}const publicKeyStr=` + strconv.Quote(key) + `;const enc=new RSAEncryptor(publicKeyStr);var cipher=enc.encrypt(u+p);$("#rsa").val(cipher);$("#loginForm")[0].submit();}`
}
