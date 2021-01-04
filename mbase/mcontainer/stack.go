/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:47:20
 * @LastEditTime: 2020-12-16 14:47:20
 * @LastEditors: Chen Long
 * @Reference:
 */

package mcontainer

type Stack struct {
	elems []interface{}
}

func (stack *Stack) New() *Stack {
	return &Stack{}
}
func (stack *Stack) Len() int {
	if stack == nil {
		return 0
	}
	return len(stack.elems)
}
func (stack *Stack) Peek() interface{} {
	l := stack.Len()
	if l == 0 {
		return nil
	}
	return stack.elems[l-1]
}
func (stack *Stack) Pop() interface{} {
	l := stack.Len()
	if l == 0 {
		return nil
	}
	v := stack.elems[l-1]
	stack.elems = stack.elems[0 : l-1]
	return v
}
func (stack *Stack) Push(elem interface{}) {
	if stack == nil {
		return
	}
	stack.elems = append(stack.elems, elem)
}
