/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-17 09:05:48
 * @LastEditTime: 2020-12-17 09:05:49
 * @LastEditors: Chen Long
 * @Reference:
 */

package mutils

import (
	"fmt"
	"strings"
)

type RulexpOperator int

const (
	RulexpOperator_EQ        RulexpOperator = 101
	RulexpOperator_EQ_DOUBLE RulexpOperator = 103
	RulexpOperator_NEQ       RulexpOperator = 105
)

type MatchFunc func(op RulexpOperator, lhs, rhs string) bool

type Rulexp interface {
	Matching(mf MatchFunc) bool
}

type OrRulexp struct {
	re1 Rulexp
	re2 Rulexp
}

func (ore *OrRulexp) Matching(mf MatchFunc) bool {
	return ore.re1.Matching(mf) || ore.re2.Matching(mf)
}
func (ore *OrRulexp) String() string {
	return fmt.Sprintf("(%v || %v)", ore.re1, ore.re2)
}

type AndRulexp struct {
	re1 Rulexp
	re2 Rulexp
}

func (are *AndRulexp) Matching(mf MatchFunc) bool {
	return are.re1.Matching(mf) && are.re2.Matching(mf)
}
func (are *AndRulexp) String() string {
	return fmt.Sprintf("(%v && %v)", are.re1, are.re2)
}

type KvRulexp struct {
	op  RulexpOperator
	key string
	val string
}

func (kvre *KvRulexp) Matching(mf MatchFunc) bool {
	return mf(kvre.op, kvre.key, kvre.val)
}
func (kvre *KvRulexp) String() string {
	if kvre == nil {
		return ""
	}
	if kvre.op == RulexpOperator_NEQ {
		return fmt.Sprintf("%s!=%s", kvre.key, kvre.val)
	} else if kvre.op == RulexpOperator_EQ_DOUBLE {
		return fmt.Sprintf("%s==%s", kvre.key, kvre.val)
	} else {
		return fmt.Sprintf("%s=%s", kvre.key, kvre.val)
	}
}

type rulexpOperator int

const (
	rulexpOperator_NONE        rulexpOperator = 0
	ruleexpOperator_R_BRACKETS rulexpOperator = 1
	rulexpOperator_OR          rulexpOperator = 3
	rulexpOperator_AND         rulexpOperator = 5
	ruleexpOperator_L_BRACKETS rulexpOperator = 7
)

func findWordTerminator(str string) int {
	terminators := []string{"=", "==", "!=", "&&", "||", ")"}
	for _, term := range terminators {
		if ind := strings.Index(str, term); ind >= 0 {
			return ind
		}
	}
	return len(str)
}
func isWordTerminator(str string) bool {
	terminators := []string{"=", "==", "!=", "&&", "||", ")"}
	for _, term := range terminators {
		if strings.HasPrefix(str, term) {
			return true
		}
	}
	return false
}
func readWord(str string) (word string, restr string, err error) {
	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return "", str, nil
	}

	prevQuota := false
	prevSlash := false
	for i, c := range str {
		if c == '\\' {
			if prevSlash {
				word += str[i : i+1]
				prevSlash = false
			} else {
				prevSlash = true
			}
		} else {
			if prevSlash {
				word += str[i : i+1]
			} else if c == '"' {
				if prevQuota {
					prevQuota = false
				} else {
					prevQuota = true
				}
			} else {
				if prevQuota || !isWordTerminator(str[i:]) {
					word += str[i : i+1]
				} else {
					return word, str[i:], nil
				}
			}
			prevSlash = false
		}
	}
	return word, "", nil
	/*
		hasQuota := false
		if str[0] == '"' {
			hasQuota = true
			str = str[1:]
		}
		strLen := len(str)
		for i := 0; i < strLen; i++ {
			if str[i] == '\\' && i+1 < strLen {

			}
		}
	*/
	/*
		if str[0] == '"' {
			lastQuota := strings.Index(str[1:], "\"")
			if lastQuota < 0 {
				return "", str, fmt.Errorf("Unexpected quotation:%s", str)
			}
			return str[1:lastQuota+1], str[1+lastQuota+1:], nil
		} else {
			last := findWordTerminator(str)
			return strings.TrimSpace(str[:last]), str[last:], nil
		}*/
}
func readKvRulexp(str string) (kre *KvRulexp, restr string, err error) {
	kre = &KvRulexp{}
	kre.key, restr, err = readWord(str)
	if err != nil {
		return nil, str, err
	}

	restr = strings.TrimSpace(restr)

	if strings.HasPrefix(restr, "==") {
		kre.op = RulexpOperator_EQ_DOUBLE
		restr = restr[2:]
	} else if strings.HasPrefix(restr, "=") {
		kre.op = RulexpOperator_EQ
		restr = restr[1:]
	} else if strings.HasPrefix(restr, "!=") {
		kre.op = RulexpOperator_NEQ
		restr = restr[2:]
	} else {
		return nil, str, fmt.Errorf("Unexpected RulexpOperator:%s", restr)
	}

	kre.val, restr, err = readWord(restr)
	if err != nil {
		return nil, str, err
	}

	return kre, restr, err
}

/*
把中缀表达式单词系列转换成后缀表达式单词系列
中缀表达式转换成后缀表达式的算法步骤：
(1).设置一个堆栈S，初始时将栈顶元素设置为#。
(2).顺序读入中缀表达式，当读到的单词为操作数时将其加入到线性表L， 并接着读下一个单词。
(3).令x1为当前栈顶运算符的变量，x2为当前扫描读到的运算符的变量，当顺序从中缀表达式中读入的单词为运算符时就赋予x2；
	然后比较x1与x2的优先级，若优先级x1<x2，将x2入栈S，接着读下一个单词；
	若优先级x1>x2，将x1从S中出栈，并加入L中，接着比较新的栈顶运算符x1与x2的优先级；
	若优先级x1=x2且x1为”(”而x2为”)”,将x1出栈，接着读下一个单词；若优先级x1=x2且x1为”#”而x2为”#”,算法结束。
————————————————
/	x2	+	-	×	÷	(	)	#
x1
+		>	>	<	<	<	>	>
-		>	>	<	<	<	>	>
×		>	>	>	>	<	>	>
÷		>	>	>	>	<	>	>
(		<	<	<	<	<	=	$
)		>	>	>	>	$	>	>
#		<	<	<	<	<	$	=
表中x1为+或-，x2为*或/时，优先级x1<x2，满足中缀表达式规则1.先乘除后加减；
x1为+、-、*或/，x2为(或/时，优先级x1<x2，满足中缀表达式规则2.先括号内后括号外；
当x1的运算符x2的运算符同级别时，优先级x1=x2，满足中缀表达式规则3.同级别时先左后右。出现表中的$表示中缀表达式语法出错。
————————————————
*/
func compareOperator(x1, x2 rulexpOperator) (r int, err error) {
	switch x1 {
	case rulexpOperator_NONE:
		if x2 == ruleexpOperator_R_BRACKETS {
			return 0, fmt.Errorf("Unexpected ')'")
		}
		return int(x1 - x2), nil
	case rulexpOperator_OR:
		return int(x1) + 1 - int(x2), nil //	+1使得同级时优先
	case rulexpOperator_AND:
		return int(x1) + 1 - int(x2), nil //	+1使得同级时优先
	case ruleexpOperator_L_BRACKETS:
		if x2 == rulexpOperator_NONE {
			return 0, fmt.Errorf("need right brackets")
		} else if x2 == ruleexpOperator_R_BRACKETS {
			return 0, nil
		} else {
			return -1, nil
		}
	case ruleexpOperator_R_BRACKETS:
		if x2 == ruleexpOperator_L_BRACKETS {
			return 0, fmt.Errorf("Unexpected '('")
		} else {
			return 1, nil
		}
	}
	return 0, fmt.Errorf("unknown rulexpOperator:%v", x1)
}
func build(ops []interface{}) (re Rulexp, err error) {
	if len(ops) == 0 {
		return nil, nil
	}

	S := make([]rulexpOperator, 0, len(ops))
	L := make([]interface{}, 0, len(ops))

	S = append(S, ops[0].(rulexpOperator))
	ops = ops[1:]

	for i, e := range ops {
		if v, ok := e.(Rulexp); ok {
			L = append(L, v)
		} else {
			x2 := e.(rulexpOperator)
			for {
				x1 := S[len(S)-1]
				r, rerr := compareOperator(x1, x2)
				if rerr != nil {
					return nil, rerr
				}
				if r < 0 {
					S = append(S, x2)
					break
				} else if r > 0 {
					L = append(L, x1)
					S = S[:len(S)-1]
				} else {
					S = S[:len(S)-1]
					if x1 == rulexpOperator_NONE && i != len(ops)-1 {
						return nil, fmt.Errorf("Unexpected none, x1=%v,x2=%v,i=%d,len(ops)=%d", x1, x2, i, len(ops))
					}
					break
				}
			}
		}
	}

	//	L已经是后缀树遍历
	res := make([]Rulexp, 0, len(L))
	for _, e := range L {
		if v, ok := e.(Rulexp); ok {
			res = append(res, v)
		} else {
			if len(res) < 2 {
				return nil, fmt.Errorf("error expression format")
			}
			switch op := e.(rulexpOperator); op {
			case rulexpOperator_OR:
				ore := &OrRulexp{re1: res[len(res)-2], re2: res[len(res)-1]}
				res = append(res[:len(res)-2], ore)
			case rulexpOperator_AND:
				ore := &AndRulexp{re1: res[len(res)-2], re2: res[len(res)-1]}
				res = append(res[:len(res)-2], ore)
			default:
				return nil, fmt.Errorf("unknown operator:%v", op)
			}
		}
	}
	if len(res) != 1 {
		return nil, fmt.Errorf("error expression")
	}
	return res[0], nil
}
func BuildRulexp(pattern string) (re Rulexp, err error) {

	//str := strings.ReplaceAll(pattern, `\"`, `"`)
	str := pattern

	var ops []interface{}
	ops = append(ops, rulexpOperator_NONE)

	for len(str) > 0 {
		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, "(") {
			ops = append(ops, ruleexpOperator_L_BRACKETS)
			str = str[1:]
		} else if strings.HasPrefix(str, ")") {
			ops = append(ops, ruleexpOperator_R_BRACKETS)
			str = str[1:]
		} else if strings.HasPrefix(str, "||") {
			ops = append(ops, rulexpOperator_OR)
			str = str[2:]
		} else if strings.HasPrefix(str, "&&") {
			ops = append(ops, rulexpOperator_AND)
			str = str[2:]
		} else {
			re, str, err = readKvRulexp(str)
			if err != nil {
				return nil, err
			}
			ops = append(ops, re)
		}
		str = strings.TrimSpace(str)
	}
	ops = append(ops, rulexpOperator_NONE)

	return build(ops)
}
