using System.Runtime.Serialization;
using YamlDotNet.Serialization;

namespace TUI.Providers.Dependencies;

[DataContract]
[YamlSerializable]
public class DependencyDto
{
    private string _icon;

    [DataMember]
    [YamlMember]
    public string? Name { get; set; }

    [DataMember]
    [YamlMember]
    public string? Icon
    {
        get => $" {_icon} ";
        set => _icon = value;
    }

    [DataMember]
    [YamlMember]
    public string? Version { get; set; }

    [DataMember]
    [YamlMember]
    public string? Color { get; set; }
}