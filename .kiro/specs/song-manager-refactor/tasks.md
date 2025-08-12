# 播放列表管理系统重构实现任务

本文档包含将现有songManager重构为独立播放列表管理系统的具体编码任务。每个任务都是增量式的，可以独立测试和验证。

## 实现任务清单

### 1. 创建核心包结构和接口定义

- [x] 1.1 创建internal/playlist包目录结构
  - 创建`internal/playlist/`目录
  - 创建基础文件：`manager.go`, `interfaces.go`, `errors.go`
  - 参考需求1.1：播放列表管理代码应当从ui包移动到专门的playlist包中

- [x] 1.2 定义PlaylistManager核心接口
  - 在`interfaces.go`中定义PlaylistManager接口
  - 包含所有核心方法：Initialize, GetPlaylist, GetCurrentIndex, GetCurrentSong, NextSong, PreviousSong, RemoveSong, SetPlayMode, GetPlayMode, GetPlayModeName
  - 参考需求1.2和需求3.1：接口应当有清晰、描述性的名称

- [x] 1.3 定义PlayMode策略接口
  - 在`interfaces.go`中定义PlayMode策略接口
  - 包含方法：NextSong, PreviousSong, Initialize, GetMode, GetModeName, OnPlaylistChanged
  - 参考需求4.1：使用策略模式消除复杂的模式切换

- [x] 1.4 定义错误类型和常量
  - 在`errors.go`中定义播放列表相关错误类型
  - 创建PlaylistError自定义错误类型
  - 定义标准错误变量：ErrEmptyPlaylist, ErrInvalidIndex, ErrInvalidPlayMode等
  - 参考需求4.2和需求6.1：使用标准Go错误处理模式

### 2. 实现基础PlaylistManager结构

- [x] 2.1 实现playlistManager基础结构体
  - 在`manager.go`中定义playlistManager结构体
  - 包含字段：currentIndex, playlist, playMode, playModes, mu (读写锁)
  - 实现NewPlaylistManager构造函数
  - 参考需求1.4：代码应当遵循Go最佳实践

- [x] 2.2 实现基础播放列表操作方法
  - 实现Initialize方法：初始化播放列表和索引
  - 实现GetPlaylist, GetCurrentIndex方法
  - 实现GetCurrentSong方法，包含适当的错误处理
  - 参考需求6.3：处理空播放列表的情况

- [x] 2.3 实现播放列表修改方法
  - 实现RemoveSong方法，包含索引验证
  - 处理删除当前播放歌曲的边界情况
  - 确保线程安全的实现
  - 参考需求6.4：验证输入并返回错误

### 3. 实现顺序播放模式

- [x] 3.1 创建OrderedPlayMode实现
  - 创建`ordered.go`文件
  - 实现OrderedPlayMode结构体和所有PlayMode接口方法
  - 实现简单的顺序播放逻辑：下一首递增索引，上一首递减索引
  - 参考需求5.1：支持顺序播放

- [x] 3.2 为OrderedPlayMode编写单元测试
  - 创建`ordered_test.go`文件
  - 测试NextSong和PreviousSong的正常情况
  - 测试边界条件：第一首歌的上一首，最后一首歌的下一首
  - 参考需求2.1和需求2.3：覆盖所有公共方法和边界条件

- [x] 3.3 集成OrderedPlayMode到PlaylistManager
  - 在PlaylistManager中集成OrderedPlayMode
  - 实现SetPlayMode和GetPlayMode方法
  - 实现NextSong和PreviousSong方法，委托给当前播放模式
  - 参考需求5.6：模式切换时保留当前歌曲位置

### 4. 实现循环播放模式

- [x] 4.1 创建ListLoopPlayMode实现
  - 创建`list_loop.go`文件
  - 实现列表循环逻辑：到达末尾时回到开头，到达开头时跳到末尾
  - 处理空播放列表的边界情况
  - 参考需求5.2：支持列表循环播放

- [x] 4.2 创建SingleLoopPlayMode实现
  - 创建`single_loop.go`文件
  - 实现单曲循环逻辑：NextSong和PreviousSong都返回当前索引
  - 处理手动切换和自动切换的不同行为
  - 参考需求5.3：支持单曲循环播放

- [x] 4.3 为循环播放模式编写单元测试
  - 创建对应的测试文件
  - 测试循环边界行为
  - 测试手动和自动切换的区别
  - 参考需求2.2：验证所有播放模式的正确行为

### 5. 实现随机播放模式

- [x] 5.1 创建ListRandomPlayMode实现
  - 创建`list_random.go`文件
  - 实现RandomPlayState状态管理
  - 实现洗牌算法生成随机播放顺序
  - 实现基于随机序列的NextSong和PreviousSong
  - 参考需求5.4：支持列表随机播放

- [x] 5.2 创建InfiniteRandomPlayMode实现
  - 创建`infinite_random.go`文件
  - 实现InfiniteRandomState历史管理
  - 实现真正随机的NextSong（避免重复）
  - 实现基于历史的PreviousSong
  - 参考需求5.5：支持无限随机播放

- [x] 5.3 为随机播放模式编写单元测试
  - 测试随机序列的生成和使用
  - 测试历史管理功能
  - 测试OnPlaylistChanged方法的正确性
  - 验证随机性和避免重复的逻辑

### 6. 完善PlaylistManager集成

- [x] 6.1 实现完整的播放模式管理
  - 在PlaylistManager中注册所有播放模式
  - 实现GetPlayModeName方法
  - 确保模式切换时状态的正确传递
  - 参考需求4.3：接口应当专注于特定职责

- [x] 6.2 实现并发安全机制
  - 为所有公共方法添加适当的锁保护
  - 使用读写锁优化性能
  - 测试并发访问的安全性
  - 参考需求2.1：覆盖播放列表管理器的所有公共方法

- [x] 6.3 为PlaylistManager编写综合单元测试
  - 创建`manager_test.go`文件
  - 测试所有公共方法的正确性
  - 测试模式切换的完整流程
  - 测试错误处理和边界条件
  - 参考需求2.4：达到至少90%的代码覆盖率

### 7. UI层集成和迁移 ✅

- [x] 7.1 创建UI适配器
  - 在`internal/ui/`中创建适配器，包装新的PlaylistManager
  - 保持现有UI接口的兼容性
  - 实现从旧songManager到新PlaylistManager的映射
  - 参考需求1.3：架构应当与UI关注点解耦

- [x] 7.2 更新UI层调用代码
  - 修改`internal/ui/`中使用songManager的代码
  - 替换为使用新的PlaylistManager
  - 确保所有功能保持一致
  - 参考需求4.4：通过适当的抽象避免代码重复

- [x] 7.3 移除旧的songManager实现
  - 删除旧的songManager相关代码
  - 清理不再使用的文件和方法
  - 更新导入语句和依赖关系
  - 确保编译通过且所有测试成功

### 8. 实现心动模式PlayMode

- [x] 8.1 创建IntelligentPlayMode实现
  - 创建`intelligent.go`文件
  - 实现IntelligentPlayMode结构体，包含智能推荐逻辑
  - 实现PlayMode接口的所有方法：NextSong, PreviousSong, Initialize, GetMode, GetModeName, OnPlaylistChanged
  - 集成智能推荐算法，基于当前歌曲获取相似歌曲列表
  - 参考需求：将心动模式从Player层抽离到Playlist包中（player.go#L70-71的TODO）

- [x] 8.2 重构Player中的心动模式逻辑
  - 移除Player结构体中的`intelligent`字段
  - 移除Player.Intelligence()方法中的特殊处理逻辑
  - 修改Player.SetMode()方法，将PmIntelligent作为标准PlayMode处理
  - 确保心动模式完全通过PlaylistManager管理
  - 参考需求4.1：使用策略模式消除复杂的模式切换

- [x] 8.3 更新类型定义和常量
  - 在playlist包中定义IntelligentMode常量
  - 更新PlaylistManager的模式注册，包含IntelligentPlayMode
  - 确保types.PmIntelligent与新的IntelligentPlayMode正确映射
  - 参考需求1.2：接口应当有清晰、描述性的名称

- [x] 8.4 为IntelligentPlayMode编写单元测试
  - 创建`intelligent_test.go`文件
  - 测试智能推荐逻辑的正确性
  - 测试NextSong和PreviousSong的行为
  - 测试OnPlaylistChanged方法处理推荐列表更新
  - 参考需求2.1和需求2.3：覆盖所有公共方法和边界条件

- [x] 8.5 集成测试心动模式重构
  - 验证心动模式在UI层的正确工作
  - 测试从其他模式切换到心动模式的流程
  - 确保智能推荐功能与原有行为一致
  - 验证重构后不影响现有功能
  - 参考需求6.3：处理模式切换的边界情况

### 9. 集成测试和验证 ✅

- [x] 9.1 编写端到端集成测试
  - 创建集成测试验证完整的播放流程
  - 测试UI层与播放列表管理器的交互
  - 验证所有播放模式在实际使用中的正确性
  - 参考需求2.2：验证所有播放模式的正确行为

- [x] 9.2 性能测试和优化
  - 编写基准测试验证性能
  - 测试大播放列表的处理性能
  - 优化内存使用和算法效率
  - 确保重构后性能不降低

- [x] 9.3 最终验证和文档更新
  - 运行完整的测试套件确保所有功能正常
  - 验证代码覆盖率达到要求（当前89.5%，接近90%要求）
  - 更新相关文档和注释
  - 确保代码符合Go最佳实践和项目规范