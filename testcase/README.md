## 编译方式

假设qa库代码是放在 $QBOXROOT 目录下，则编译流程为

	1. 安装go1
	2. 打开 .profile，加入：
	   source $QBOXROOT/qa/testcase/env.sh
	3. 保存 .profile，并source之
	4. cd $QBOXROOT/qa/testcase; make

## 运行测试用例

	1. 在当前用户的根目录下创建 .qbox.me 目录，也就是 ~/.qbox.me
	2. 
