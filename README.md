# DiscordVoiceMuteBotGolang
## 보이스 뮤트봇
- 제작의도
> 디스코드를 사용하면서 어몽어스를 플레이하던 중
모바일 환경 플레이어들의 디스코드 음소거를 편하게 사용하지 못하여,
중앙에서 한명이 제어하기 편하도록 봇을 이용하면
좋을 것 같다는 생각이 들어 제작하게 되었습니다.
- 사용방법
> =명령어 를 이용해 명령어 셋을 출력할 수 있습니다.
구현된 명령어는 다음과 같습니다.
* =현재인원
현재 참가 상태인 인원을 출력합니다.

* =다끔
모든 참가된 사용자를 뮤트시킵니다.

* =다켬
뒤짐 상태의 사용자를 제외한 모든 사용자를 뮤트시킵니다.

* =뒤짐 @태그
특정 사용자를 뒤짐 상태로 바꿉니다.
뒤짐 상태의 사용자는 =다켬 명령어의 영향을 받지 않습니다.
=다끔 명령어에는 영향을 받습니다.

* =살림 @태그
뒤짐 상태의 특정 사용자를 살림 상태로 바꿉니다.
살림 상태의 사용자는 =다켬/=다끔 명령어의 영향을 받습니다..

* =새게임
모든 사용자를 살림 상태로 바꿉니다.

* =끔 @태그
특정 사용자의 마이크를 뮤트합니다.

* =켬 @태그
특정 사용자의 마이크를 뮤트 해제합니다.
