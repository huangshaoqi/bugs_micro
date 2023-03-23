// 流水线（声明式）
pipline{

	// 任何一个代理可用就可以执行
	angent any

	// 定义一些环境信息

	// 定义流水线的加工流程
	stages {
		// 流水线的所有阶段
		// 1.编译 “bugs_micro项目”
		stage('代码编译'){
			steps{
				// 要做的所有事情
				echo "1.编译。。。"
			}
		}
		
		stage('测试'){
			steps{
				// 要做的所有事情
				echo "2.测试。。。"
			}
		}
		stage('打包'){
			steps{
				// 要做的所有事情
				echo "3.打包。。。"
			}
		}
		stage('部署'){
			steps{
				// 要做的所有事情
				echo "4.部署。。"
			}
		}
	}
}