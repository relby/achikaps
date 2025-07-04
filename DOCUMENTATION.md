# Документация

### Модели
- Нода (Node)
```json
{
    "ID": uint
    "SessionID": string
    "Type": uint // 1 - Transit, 2 - Production
    "Name": uint
    "Position": {"X": float64, "Y": float64}
    "Radius": float64
    "BuildProgress": float64 // Значение от 0 до 1, если 1 то нода построена
}
```

- Юнит (Unit)
```json
{
    "Type": uint // 1 - Idle, 2 - Production, 3 - Builder, 4 - Transport
    "SessionID": string
    "Node": Node // см. выше
    "Material": Material // только для Transport типа
    "Action": UnitAction // см. ниже
}
```

- Действие юнита (UnitAction)
```json
{
    "Type": // 1 - Moving, 2 - Production, 3 - Building, 4 - Transport
    "IsStarted": bool
    "Data": any // Данные, зависящие от типа
}

// MovingUnitActionData
{
    "Speed": float64
    "TimeMs": float64
    "FromNode": Node // От этой ноды юнит начал движение
    "ToNode": Node // К этой ноде движется юнит
    "Progress": float64 // Значение от 0 до 1, обозначающее продвижение по дороге от одной ноде к другой
}

// ProductionUnitActionData 
{
    "Progress": float64
}

// TakeMaterialUnitActionData
{
    "Material": Material
}
// DropMaterialUnitActionData
{}
// BuildingUnitActionData
{}

```

- Материал (Material)
```json
{
    "ID": uint
    "SessionID": string
    "Type": uint
    "Node": Node
    "IsReserved": bool
}
```
- Типы Материалов (MaterialType)
1. `GrassMaterialType`
2. `SandMaterialType`
3. `DewMaterialType`
4. `SeedMaterialType`
5. `SugarMaterialType`
6. `JuiceMaterialType`
7. `ChitinMaterialType`
8. `EggMaterialType`
9. `PheromoneMaterialType`
10. `AmberMaterialType`

- Условие победы (WinCondition)
```json
{
    "MaterialType": MaterialType
    "Count": int
}
```

### Оп коды
- 1. Получение стартого стэйта
  - Ответ:
    ```json
    {
        "Nodes": Map<SessionID, Map<NodeID, Node>>
        "Connections": Map<SessionID, Map<NodeID, List<NodeID>>>
        "Units": Map<SessionID, Map<UnitID, Unit>>
        "Materials": Map<SessionID, Map<MaterialID, Material>>
        "WinCondition": WinCondition
    }
    ```
- 2. Строительство ноды
  - Запрос:
    ```json
    {
        "FromNodeID": uint // От какой ноды строить путь до новой
        "Name": uint // Названия ноды, описан ниже
        "Position": {"X": float64, "Y": float64} // Позиция новой ноды
    }
    ```
  - Названия нод:
    1. `SandTransitNodeName`
    2. `GrassFieldNodeName`
    3. `WellNodeName`
    4. `SeedStorageNodeName`
    5. `AphidDistillationNodeName`
    6. `RawMaterialVatNodeName`
    7. `ChitinPressNodeName`
    8. `EggFarmNodeName`
    9. `PheromoneMineNodeName`
    10. `IncubatorNodeName`
    11. `GeneticHatcheryNodeName`
    12. `GuardOutpostNodeName`
    13. `AmberTurretNodeName`
  - Ответа:
    1. Успех: 
    ```json
    {
        "FromNodeID": uint
        "Node": Node
    }
    ```
    2. Ошибка: `{"error": string}`
- 3. Начало выполнение действия юнитом
  - Ответ: `Map<SessionID, UnitActionExecuteResp>`

  Модель `UnitActionExecuteResp`
  ```json
  {
    "Unit": Unit
    "UnitAction": UnitAction
  }
  ```
- 4. Изменение типа юнита
  - Запрос:
  ```json
  {
    "ID": uint
    "Type": uint
  }
  ```
  - Ответ:
    1. Успех: 
    ```json
    {
        "Unit": Unit
    }
    ```
    2. Ошибка: `{"error": string}`
- 5. Победа одного из игроков
  - Ответ: `WinResp`

  Модель `WinResp`
  ```json
  {
    "SessionID": string
  }
  ```
- 6. Постройка ноды
  - Ответ: `NodeBuiltResp`

  Модель `NodeBuiltResp`
  ```json
  {
    "Node": Node
  }
  ```
- 7. Удаление материала
  - Ответ: `MaterialDestroyedResp`

  Модель `MaterialDestroyedResp`
  ```json
  {
    "Material": Material
  }
  ```
- 8. Создание материала
  - Ответ: `MaterialCreatedResp`

  Модель `MaterialCreatedResp`
  ```json
  {
    "Material": Material
  }
  ```
- 9. Создание юнита
  - Ответ: `UnitCreatedResp`

  Модель `UnitCreatedResp`
  ```json
  {
    "Unit": Unit
  }
  ```