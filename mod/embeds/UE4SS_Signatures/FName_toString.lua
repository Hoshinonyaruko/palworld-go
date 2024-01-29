function Register()
  return "48 89 5C 24 10 48 89 74 24 18 57 48 83 EC 20 80 3D ? ? ? ? ? 48 8B FA 8B 19 48 8B F1 74 09 48 8D ? ? ? ? ? EB 16 48 8D ? ? ? ? ? E8 ? ? ? ? 48 8B D0 C6 05 ? ? ? ? ? 8B CB 0F B7 C3 C1 E9 10 89 4C 24 30 89 44 24 34 48 8B 44 24 30 48 C1 E8 20 8D 1C 00 48 03 5C CA 10 48 8B CF"
end

function OnMatchFound(MatchAddress)
  return MatchAddress
end