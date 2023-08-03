using YamlDotNet.Serialization;


namespace TUI.Domain;

[YamlSerializable]
public class Source
{
    [YamlMember]
    public string[] Tags { get; set; }

    [YamlMember]
    public string Name { get; set; }

    [YamlMember]
    public string Repo { get; set; }
}