# Локальные источники

В колонке «Источник» указаны ревизии для проверки происхождения и планового обновления. Для выполнения инструкций используются локальные файлы. Лицензии сторонних авторов сохранены рядом с комплектом.

| Локальный материал | Источник и ревизия | Лицензия |
|---|---|---|
| [План](skills/writing-plans/SKILL.md), [запрос ревью](skills/requesting-code-review/SKILL.md), [разбор замечаний](skills/receiving-code-review/SKILL.md), [TDD](skills/test-driven-development/SKILL.md), [проверка результата](skills/verification-before-completion/SKILL.md), [рабочие деревья](skills/using-git-worktrees/SKILL.md), [завершение ветки](skills/finishing-a-development-branch/SKILL.md) | [obra/superpowers, каталоги skills с теми же именами](https://github.com/obra/superpowers/tree/22f0b463e7a5e579904dd6211a520321035526f7/skills) | [MIT](licenses/superpowers.LICENSE) |
| [Подготовка коммита](skills/git-commit/SKILL.md) | [github/awesome-copilot, git-commit](https://github.com/github/awesome-copilot/tree/f10fdb8b2660ce85c98f220b2f1ca24653768a5c/skills/git-commit) | [MIT](licenses/github-awesome-copilot.LICENSE) |
| [Тестирование Go](skills/golang-testing/SKILL.md) | [affaan-m/ECC, golang-testing](https://github.com/affaan-m/ECC/tree/4adeee6732a12d1540f49a4d50a38ce8c97f9a53/skills/golang-testing) | [MIT](licenses/ecc.LICENSE) |
| [Браузерные проверки](skills/webapp-testing/SKILL.md) | [anthropics/skills, webapp-testing](https://github.com/anthropics/skills/tree/e4bb5d960d87126ff3b3a8af1a104b7db67560ba/skills/webapp-testing) | [Apache-2.0](skills/webapp-testing/LICENSE.txt) |
| [Ревью границ доверия](skills/differential-review/SKILL.md) | [Trail of Bits, differential-review](https://github.com/trailofbits/skills/tree/149d7b21f4067765b82e67029730b42d228ef4b4/plugins/differential-review/skills/differential-review) | [CC BY-SA 4.0](licenses/trailofbits-skills.LICENSE), включая русскую адаптацию |

## Использование и адаптации

- В writing-plans маршрут исполнения заменён на локальный [порядок делегирования](../AGENTS.md). Имена веток, русский язык сообщений, авторство и условия интеграции определяются [правилами Git](git.md). Примеры сторонних авторов служат техническими иллюстрациями.
- Методика Trail of Bits сокращена и переработана в русскую памятку защитного ревью. Для Go выбран материал ECC с явной лицензией. Режимы инструментов, версии библиотек и цели покрытия выбирать по требованиям проекта.
- В finishing-a-development-branch маршрут интеграции заменён локальной процедурой PR и squash. В инструкциях для коммитов и worktree добавлена ссылка на местные правила.
- Браузерный скилл адаптирован для выделенного стенда с проверкой идентификатора сборки. Запуском и освобождением среды управляет её владелец. Примеры используют переменные окружения и локальные пути вывода.
- Команды Git требуют Git, Go-команды — версию Go из проекта, браузерные сценарии — Python, Playwright и совместимый браузер. Подготовить эти зависимости перед запуском соответствующих проверок. Для чтения комплекта сеть и установка расширений не требуются.
- При обновлении сверить исходный каталог, лицензию, сопровождающие файлы и локальные адаптации. Записать новый SHA, проверить ссылки и направить изменение на независимое ревью.
