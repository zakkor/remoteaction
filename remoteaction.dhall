let Auth = {
  Type = {
    username: Text,
    password: Text
  },
  default = {
    username = "",
    password = ""
  }
 }

let Pattern = {
  Type = {
    url : Text,
    regex : Text
  },
  default = {
    url = "",
    regex = ""
  }
}

let Action : Type = {
  name : Text,
  script : Text
}

let LinkMenu : Type = {
  patterns : List Pattern.Type,
  actions : List Action
}

let SelectionMenu : Type = {
  regexes : List Text,
  actions : List Action
}

let Menu = {
  Type = {
    link : Optional LinkMenu,
    selection : Optional SelectionMenu
  },
  default = {
    link = None LinkMenu,
    selection = None SelectionMenu
  }
}

let Config = {
  Type = {
      listen : Text,
      auth : Auth.Type,
      executors : Natural,
      menus : List Menu.Type

  },
  default = {
    listen = "localhost:6969",
    auth = Auth::{=},
    executors = 5
  }
}

in {
  Auth = Auth,
  Pattern = Pattern,
  Action = Action,
  LinkMenu = LinkMenu,
  SelectionMenu = SelectionMenu,
  Menu = Menu,
  Config = Config,
  Link = \(menu : LinkMenu) -> Menu::{ link = Some menu },
  Selection = \(menu : SelectionMenu) -> Menu::{ selection = Some menu },
  regex = \(pattern : Text) -> Pattern::{ regex = pattern },
  url = \(pattern : Text) -> Pattern::{ url = pattern }
}
