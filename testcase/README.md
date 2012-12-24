## 编译方式

假设qa库代码是放在 $QBOXROOT 目录下，则编译流程为

	1. 安装go1
	2. 打开 .profile，加入：
	   source $QBOXROOT/qa/testcase/env.sh
	3. 保存 .profile，并source之
	4. cd $QBOXROOT/qa/testcase; make

## 运行测试用例

	1. cd $QBOXROOT/qa/testcase; make install
	2. 用例配置放在$QBOXROOT/qa/testcase/testing/conf.d目录下，包括cases配置文件，数据文件和执行环境env
	3. 运行 qboxtestcase
