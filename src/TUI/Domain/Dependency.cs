using YamlDotNet.Serialization;


namespace TUI.Domain;

[YamlSerializable]
public class Dependency
{
    private string _icon;

    [YamlMember]
    public string Name { get; set; }

    [YamlMember]
    public string Icon
    {
        get => $" {_icon} ";
        set => _icon = value;
    }

    [YamlMember]
    public string Version { get; set; }

    [YamlMember]
    public string Color { get; set; }
}